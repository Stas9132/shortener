package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"net/http"
	"shortener/config"
	"shortener/internal/app/handlers"
	"shortener/internal/logger"
	"sync"
)

var r = sync.OnceValue(func() *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)

	r.Post("/", handlers.MainPage)
	r.Get("/{sn}", handlers.GetByShortName)
	r.Post("/api/shorten", handlers.JSONHandler)
	r.NotFound(handlers.Default)
	r.MethodNotAllowed(handlers.Default)
	return r
})

func run() error {
	if err := logger.Initialize(*config.LogLevel); err != nil {
		return err
	}
	logger.Log.WithFields(logrus.Fields{
		"address": *config.ServerAddress,
	}).Infoln("Starting server")
	return http.ListenAndServe(*config.ServerAddress, r())
}

func main() {
	config.InitConfig()

	if err := run(); err != nil {
		panic(err)
	}
}
