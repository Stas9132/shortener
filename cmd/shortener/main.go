package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"shortener/internal/app/handlers"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", handlers.MainPage)
	r.Get("/{sn}", handlers.GetByShortName)
	r.NotFound(handlers.Default)
	r.MethodNotAllowed(handlers.Default)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
