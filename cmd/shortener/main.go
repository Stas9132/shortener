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
	"sync"
	"time"
)

var server = sync.OnceValue(func() *http.Server {
	return &http.Server{
		Addr:    *config.ServerAddress,
		Handler: r,
	}
})
var r *chi.Mux

func mRouter(handler handlers.APII) {
	r = chi.NewRouter()
	r.Use(middlware.RequestLogger, middlware.Authorization, gzip.GzipMiddleware)

	r.Post("/", handler.PostPlainText)
	r.Get("/{sn}", handler.GetRoot)
	r.Post("/api/shorten", handler.PostJSON)
	r.Post("/api/shorten/batch", handler.PostBatch)
	r.Get("/api/user/urls", handler.GetUserURLs)
	r.Get("/ping", handler.GetPing)
	r.NotFound(handler.Default)
	r.MethodNotAllowed(handler.Default)
	//http.Handle("/", r)
}

func run(h handlers.APII) {
	logger.WithFields(map[string]interface{}{
		"address": *config.ServerAddress,
	}).Infoln("Starting server")

	mRouter(h)

	if err := server().ListenAndServe(); err != nil {
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
	l := logger.NewLogger(ctx)
	var st handlers.StorageI
	if len(*config.DatabaseDsn) == 0 {
		st = storage.NewFileStorage(ctx, l)
	} else {
		st = storage.NewDB(ctx, l)
	}
	h := handlers.NewAPI(ctx, l, st)
	go run(h)

	<-ctx.Done()

	ctx2, can2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer can2()
	server().Shutdown(ctx2)
	st.Close()
	time.Sleep(time.Second)
}
