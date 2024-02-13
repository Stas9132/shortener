package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/Stas9132/shortener/config"
	"github.com/Stas9132/shortener/internal/app/handlers"
	"github.com/Stas9132/shortener/internal/app/handlers/middleware"
	"github.com/Stas9132/shortener/internal/app/model"
	"github.com/Stas9132/shortener/internal/app/proto"
	"github.com/Stas9132/shortener/internal/app/storage"
	"github.com/Stas9132/shortener/internal/gzip"
	"github.com/Stas9132/shortener/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)
}

func mRouter(handler handlers.APII) {
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger, middleware.Authorization, gzip.GzipMiddleware)

	r.Post("/", handler.PostPlainText)
	r.Get("/{sn}", handler.GetRoot)
	r.Post("/api/shorten", handler.PostJSON)
	r.Post("/api/shorten/batch", handler.PostBatch)
	r.Get("/api/user/urls", handler.GetUserURLs)
	r.Delete("/api/user/urls", handler.DeleteUserUrls)
	r.Get("/api/internal/stats", handler.GetStats)
	r.Get("/ping", handler.GetPing)
	r.NotFound(handler.Default)
	r.MethodNotAllowed(handler.Default)
	http.Handle("/", r)
}

func run(s *http.Server, h handlers.APII) {
	listenSrv := func(f any, parms ...string) {
		var err error
		switch t := f.(type) {
		case func(string, string) error:
			err = t(parms[0], parms[1])
		case func() error:
			err = t()
		}
		if err != nil {
			t := &net.OpError{}
			if errors.As(err, &t) {
				log.Fatal(err)
			} else {
				log.Println(err)
			}
		}
	}

	logger.WithFields(map[string]interface{}{
		"address": config.C.ServerAddress,
	}).Infoln("Starting server")

	mRouter(h)

	if config.C.SecureConnection {
		listenSrv(s.ListenAndServeTLS, "server.crt", "server.key")
	} else {
		listenSrv(s.ListenAndServe)
	}
}

func getServerOptions() (opt []grpc.ServerOption) {

	if config.C.SecureConnection {
		tlsSert, _ := tls.LoadX509KeyPair("server.crt", "server.key")
		tlsCfg := &tls.Config{Certificates: []tls.Certificate{tlsSert}, ClientAuth: tls.NoClientCert}
		opt = append(opt, grpc.Creds(credentials.NewTLS(tlsCfg)))
	} else {
		opt = append(opt, grpc.Creds(insecure.NewCredentials()))
	}
	return
}

func runGRPC(s *grpc.Server, l logger.Logger) {
	logger.WithFields(map[string]interface{}{
		"address": config.C.ServerAddressGRPC,
	}).Infoln("Starting grpc server")

	listen, err := net.Listen("tcp", config.C.ServerAddressGRPC)
	if err != nil {
		log.Fatal(err)
	}
	proto.RegisterApiServer(s, handlers.NewGRPCAPI(l))
	reflection.Register(s)

	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	config.Init(ctx)

	l, err := logger.NewLogrusLogger(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var st handlers.StorageI
	if len(config.C.DatabaseDsn) == 0 {
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
	m := model.NewAPI(l, st)
	h := handlers.NewAPI(ctx, l, st, m)
	s := &http.Server{Addr: config.C.ServerAddress}
	g := grpc.NewServer(getServerOptions()...)
	go run(s, h)
	go runGRPC(g, l)

	<-ctx.Done()

	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cansel()
	s.Shutdown(ctx)
	g.GracefulStop()
	st.Close()
}
