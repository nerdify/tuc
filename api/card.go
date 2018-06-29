package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/apex/log"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/tj/go/env"
	"github.com/tj/go/http/response"

	"github.com/nerdify/tuc"
	"github.com/nerdify/tuc/client"
)

var c = client.NewClient(env.Get("ENDPOINT"))
var cache = gocache.New(5*time.Minute, 10*time.Minute)
var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
		log.Error(err)
		response.Unauthorized(w)
	},
	SigningMethod: jwt.SigningMethodHS256,
	UserProperty:  "token",
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte(env.Get("JWT_KEY")), nil
	},
})

// CardHandler handles communication with the Card related methods.
type CardHandler struct {
	CardService tuc.CardService
}

// NewCardHandler returns a new instance of CardHandler.
func NewCardHandler(r *mux.Router) *CardHandler {
	h := &CardHandler{}

	s := r.NewRoute().Subrouter()
	s.Use(jwtMiddleware.Handler)
	s.HandleFunc("/cards", h.handleGetCards).Methods(http.MethodGet)
	s.HandleFunc("/cards", h.handlePostCard).Methods(http.MethodPost)
	s.HandleFunc("/cards/{card}", h.handleDeleteCard).Methods(http.MethodDelete)
	s.HandleFunc("/cards/{card}/balance", h.handleGetCardBalance).Methods(http.MethodGet)

	return h
}

func (h *CardHandler) handleGetCards(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	cards, err := h.CardService.List(userID)

	if err != nil {
		log.WithError(err).Error("loading cards")
		response.InternalServerError(w)
		return
	}

	response.OK(w, cards)
}

func (h *CardHandler) handlePostCard(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name   string `json:"name"`
		Number string `json:"number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.WithError(err).Error("parsing body")
		response.BadRequest(w)
		return
	}

	l := log.WithFields(log.Fields{
		"name":   body.Name,
		"number": body.Number,
	})

	if err := validateCard(body.Name, body.Number); err != nil {
		l.Error("invalid request")
		response.JSON(w, map[string]string{"message": err.Error()}, http.StatusUnprocessableEntity)
		return
	}

	out, err := c.GetBalance(&client.RequestInput{
		Card: body.Number,
	})

	if err != nil {
		log.WithError(err).Error("making request")
		response.InternalServerError(w)
		return
	}

	if out.Code == 2 {
		l.Warn("card does not exist")
		response.NotFound(w)
		return
	}

	data := out.Data[0]

	if strings.ToLower(data.Status) == "bloqueado" {
		l.Warn("inactive card")
		response.BadRequest(w)
		return
	}

	card := &tuc.Card{
		Balance: data.Balance,
		ID:      uuid.NewV4().String(),
		Name:    body.Name,
		Number:  body.Number,
		UserID:  getUserID(r),
	}

	if err := h.CardService.Create(card); err != nil {
		l.WithError(err).Error("creating card")
		response.InternalServerError(w)
		return
	}

	response.Created(w, card)
}

func (h *CardHandler) handleDeleteCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getUserID(r)

	if err := h.CardService.Delete(userID, vars["card"]); err != nil {
		log.WithError(err).Error("deleting card")
		response.InternalServerError(w)
		return
	}

	response.NoContent(w)
}

func (h *CardHandler) handleGetCardBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardID := vars["card"]
	userID := getUserID(r)

	l := log.WithField("card", cardID)

	card, err := h.CardService.Get(userID, cardID)

	if err != nil {
		l.WithError(err).Error("loading card")
		response.InternalServerError(w)
		return
	}

	if card == nil {
		l.Warn("card does not exist")
		response.NotFound(w)
		return
	}

	cacheKey := "tuc:" + card.Number

	// get from cache
	if balance, found := cache.Get(cacheKey); found {
		response.OK(w, map[string]interface{}{
			"balance": balance,
		})
		return
	}

	out, err := c.GetBalance(&client.RequestInput{
		Card: card.Number,
	})

	if err != nil {
		l.WithError(err).Error("making request")
		response.InternalServerError(w)
		return
	}

	if out.Code == 2 {
		l.Warn("card does not exist")
		response.NotFound(w)
		return
	}

	data := out.Data[0]

	if strings.ToLower(data.Status) == "bloqueado" {
		l.Warn("inactive card")
		response.BadRequest(w)
		return
	}

	balance := data.Balance

	if _, err := h.CardService.Update(userID, cardID, balance); err != nil {
		l.WithError(err).Error("updating card")
		response.InternalServerError(w)
		return
	}

	// set to cache
	cache.SetDefault(cacheKey, balance)

	response.OK(w, map[string]interface{}{
		"balance": balance,
	})
}

func getUserID(r *http.Request) string {
	token := r.Context().Value("token").(*jwt.Token)

	return token.Claims.(jwt.MapClaims)["jti"].(string)
}

func validateCard(name, number string) error {
	if name == "" {
		return errors.New("El nombre es requerido")
	} else if m, _ := regexp.MatchString("^\\d{8}$", number); !m {
		return errors.New("El número debe ser de 8 dígitos")
	}

	return nil
}
