package main

import (
	"net/http"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	texthandler "github.com/apex/log/handlers/text"
	"github.com/gorilla/mux"
	"github.com/nerdify/tuc/dynamodb"
	"github.com/tj/go/env"

	"github.com/nerdify/tuc/api"
)

func init() {
	if env.GetDefault("UP_STAGE", "development") == "development" {
		log.SetHandler(texthandler.Default)
	} else {
		log.SetHandler(jsonhandler.Default)
	}
}

func main() {
	addr := ":" + env.Get("PORT")

	http.Handle("/", buildRouter())

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Fatal("binding")
	}
}

func buildRouter() *mux.Router {
	app := mux.NewRouter().PathPrefix("/api").Subrouter()

	uh := api.NewAuthHandler(app)
	uh.UserService = &dynamodb.UserService{}
	uh.LoginRequestService = &dynamodb.LoginRequestService{}

	ch := api.NewCardHandler(app)
	ch.CardService = &dynamodb.CardService{}

	return app
}
