package main

import (
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"shortener/config"
	"shortener/internal/app/handlers"
)

func main() {
	config.ServerAddress = flag.String("a", "localhost:8080", "Address of http server")
	config.ResponsePrefix = flag.String("b", "http://localhost:8080/", "Response prefix")
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", handlers.MainPage)
	r.Get("/{sn}", handlers.GetByShortName)
	r.NotFound(handlers.Default)
	r.MethodNotAllowed(handlers.Default)

	log.Println("Starting server on", *config.ServerAddress)
	err := http.ListenAndServe(*config.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
