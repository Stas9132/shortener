package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortener/config"
	"shortener/internal/app/handlers"
	"shortener/internal/app/handlers/middlware"
	"shortener/internal/app/storage"
	"shortener/internal/gzip"
	"shortener/internal/logger"
	"sync"
	"time"
)

var server = sync.OnceValue(func() *http.Server {
	return &http.Server{
		Addr: *config.ServerAddress,
	}
})

func mRouter(handler handlers.ApiI) {
	r := chi.NewRouter()
	r.Use(middlware.RequestLogger, gzip.GzipMiddleware)

	r.Post("/", handler.PostPlainText)
	r.Get("/{sn}", handler.GetRoot)
	r.Post("/api/shorten", handler.PostJSON)
	r.Get("/api/user/urls", handler.GetUserURLs)
	r.NotFound(handler.Default)
	r.MethodNotAllowed(handler.Default)
	http.Handle("/", r)
}

func run(h handlers.ApiI) {
	logger.WithFields(map[string]interface{}{
		"address": *config.ServerAddress,
	}).Infoln("Starting server")

	mRouter(h)

	if err := server().ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config.Init()
	logger.Init()
	st := storage.New()
	h := handlers.NewApi(st)
	go run(h)

	<-ctx.Done()

	ctx2, can2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer can2()
	server().Shutdown(ctx2)
	st.Close()
}
