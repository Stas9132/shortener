package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"shortener/config"
	"shortener/internal/app/handlers"
)

func main() {
	config.InitConfig()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", handlers.MainPage)
	r.Get("/{sn}", handlers.GetByShortName)
	r.Post("/api/shorten", handlers.JSONHandler)
	r.NotFound(handlers.Default)
	r.MethodNotAllowed(handlers.Default)

	log.Println("Starting server on", *config.ServerAddress)
	err := http.ListenAndServe(*config.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
