package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"shortener/config"
	"shortener/internal/app/handlers"
	"shortener/internal/app/handlers/middlware"
	"shortener/internal/app/storage"
	"shortener/internal/gzip"
	"shortener/internal/logger"
	"time"
)

func mRouter(handler handlers.APII) {
	r := chi.NewRouter()
	r.Use(middlware.RequestLogger, middlware.Authorization, gzip.GzipMiddleware)

	r.Post("/", handler.PostPlainText)
	r.Get("/{sn}", handler.GetRoot)
	r.Post("/api/shorten", handler.PostJSON)
	r.Post("/api/shorten/batch", handler.PostBatch)
	r.Get("/api/user/urls", handler.GetUserURLs)
	r.Delete("/api/user/urls", handler.DeleteUserUrls)
	r.Get("/ping", handler.GetPing)
	r.NotFound(handler.Default)
	r.MethodNotAllowed(handler.Default)
	http.Handle("/", r)
}

func run(s *http.Server, h handlers.APII) {
	logger.WithFields(map[string]interface{}{
		"address": *config.ServerAddress,
	}).Infoln("Starting server")

	mRouter(h)

	if err := s.ListenAndServe(); err != nil {
		t := &net.OpError{}
		if errors.As(err, &t) {
			log.Fatal(err)
		} else {
			log.Println(err)
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config.Init(ctx)
	l, err := logger.NewLogger(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var st handlers.StorageI
	if len(*config.DatabaseDsn) == 0 {
		st, err = storage.NewFileStorage(ctx, l)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		st, err = storage.NewDB(ctx, l)
		if err != nil {
			log.Fatal(err)
		}
	}
	h := handlers.NewAPI(ctx, l, st)
	s := &http.Server{Addr: *config.ServerAddress}
	go run(s, h)

	<-ctx.Done()

	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cansel()
	s.Shutdown(ctx)
	st.Close()
	time.Sleep(time.Second)
}
