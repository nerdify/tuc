package main

import (
	"net/http"
	"strings"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/bmizerany/pat"
	"github.com/tj/go/env"
	"github.com/tj/go/http/response"

	"github.com/hosmelq/tuc-balance/internal/cache"
	"github.com/hosmelq/tuc-balance/internal/client"
)

var c = client.Client{
	Endpoint: env.Get("ENDPOINT"),
	Token:    env.Get("TOKEN"),
}

func init() {
	log.SetHandler(jsonhandler.Default)
}

func main() {
	app := pat.New()

	app.Get("/:card", http.HandlerFunc(getCardBalance))

	addr := ":" + env.Get("PORT")

	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("binding")
	}
}

func getCardBalance(w http.ResponseWriter, r *http.Request) {
	card := r.URL.Query().Get(":card")
	cacheKey := "tuc:" + card

	obj := make(map[string]interface{})

	if err := cache.Get(cacheKey, &obj); err == nil {
		response.OK(w, obj)

		return
	}

	out, err := c.GetBalance(&client.RequestInput{
		Card: card,
	})

	if err != nil {
		log.Fatalf("error making request: %s", err)
	}

	if out.StatusCode >= 400 {
		if strings.ToLower(out.Message) == "inactivo" {
			response.BadRequest(w, out.StatusCode)
		} else {
			response.NotFound(w)
		}

		return
	}

	obj["balance"] = out.Balance
	obj["number"] = card

	cache.Set(cacheKey, obj)

	response.OK(w, obj)
}