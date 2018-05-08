package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/tj/go/env"
	"github.com/tj/go/http/response"

	"github.com/nerdify/tuc"
)

var views = packr.NewBox("./views")

// AuthHandler handles communication with the Auth related methods.
type AuthHandler struct {
	UserService         tuc.UserService
	LoginRequestService tuc.LoginRequestService
}

// NewAuthHandler returns a new instance of AuthHandler.
func NewAuthHandler(r *mux.Router) *AuthHandler {
	h := &AuthHandler{}

	r.HandleFunc("/login", h.handleLogin).Methods(http.MethodPost)
	r.HandleFunc("/login/facebook", h.handleFacebookLogin).Methods(http.MethodPost)
	r.HandleFunc("/authenticate", h.handleAuthenticate).Methods(http.MethodGet)
	r.HandleFunc("/access_token", h.handleAccessToken).Methods(http.MethodPost)

	return h
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.WithError(err).Error("parsing body")
		response.BadRequest(w)
		return
	}

	if body.Email == "" {
		log.Error("invalid email")
		response.BadRequest(w)
		return
	}

	u, err := h.UserService.Find(body.Email)

	if err != nil {
		log.WithError(err).Error("loading user")
		response.InternalServerError(w)
		return
	}

	if u == nil {
		u = &tuc.User{
			ID: body.Email,
		}

		if err := h.UserService.Create(u); err != nil {
			log.WithError(err).Error("creating user")
			response.InternalServerError(w)
			return
		}
	}

	v := tuc.LoginRequest{
		RequestToken:      uuid.NewV4().String(),
		UserID:            u.ID,
		VerificationToken: uuid.NewV4().String(),
		Verified:          false,
	}

	if err = h.LoginRequestService.Create(&v); err != nil {
		log.WithError(err).Error("creating login request")
		response.InternalServerError(w)
		return
	}

	response.OK(w, map[string]string{
		"code": v.RequestToken,
	})
}

func (h *AuthHandler) handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.WithError(err).Error("parsing body")
		response.BadRequest(w)
		return
	}

	res, err := http.Get("https://graph.facebook.com/me?fields=email,id&access_token=" + body.AccessToken)
	if err != nil {
		log.WithError(err).Error("requesting facebook permissions")
		response.InternalServerError(w)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Error("invalid access token")
		response.BadRequest(w)
		return
	}

	var fbr struct {
		Email string `json:"email"`
		ID    string `json:"id"`
	}

	if err := json.NewDecoder(res.Body).Decode(&fbr); err != nil {
		log.WithError(err).Error("parsing facebook response")
		response.InternalServerError(w)
		return
	}

	if fbr.Email == "" {
		response.Unauthorized(w)
		return
	}

	u, err := h.UserService.Find(fbr.Email)
	if err != nil {
		log.WithError(err).Error("loading user")
		response.InternalServerError(w)
		return
	}

	if u == nil {
		u = &tuc.User{
			FacebookID: fbr.ID,
			ID:         fbr.Email,
		}

		if err := h.UserService.Create(u); err != nil {
			log.WithError(err).Error("creating user")
			response.InternalServerError(w)
			return
		}
	} else if u.FacebookID == "" {
		u.FacebookID = fbr.ID

		if err := h.UserService.Update(u); err != nil {
			log.WithError(err).Error("updating item")
			response.InternalServerError(w)
			return
		}
	}

	token, err := generateAccessToken(fbr.Email)

	if err != nil {
		log.WithError(err).Error("signed token")
		response.InternalServerError(w)
		return
	}

	response.OK(w, map[string]string{
		"email": u.ID,
		"token": token,
	})
}

func (h *AuthHandler) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	token := r.URL.Query().Get("token")

	if email == "" || token == "" {
		response.BadRequest(w)
		return
	}

	if err := h.LoginRequestService.Verify(email, token); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.WithError(aerr).Error("condition failed")
				response.Unauthorized(w)
				return
			}
		}

		log.WithError(err).Error("authenticating")
		response.InternalServerError(w)
		return
	}

	t := template.Must(template.New("").Parse(views.String("authenticate.html")))

	t.Execute(w, map[string]string{})
}

func (h *AuthHandler) handleAccessToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code  string `json:"code"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.WithError(err).Error("parsing body")
		response.BadRequest(w)
		return
	}

	if err := h.LoginRequestService.Delete(body.Email, body.Code); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.WithError(aerr).Error("condition failed")
				response.Unauthorized(w)
				return
			}
		}

		log.WithError(err).Error("deleting request")
		response.BadRequest(w)
		return
	}

	token, err := generateAccessToken(body.Email)

	if err != nil {
		log.WithError(err).Error("signed token")
		response.InternalServerError(w)
		return
	}

	response.OK(w, map[string]string{
		"token": token,
	})
}

func generateAccessToken(email string) (string, error) {
	key := []byte(env.Get("JWT_KEY"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:       email,
		IssuedAt: time.Now().Unix(),
		Issuer:   "saldotuc.com",
	})

	return token.SignedString(key)
}
