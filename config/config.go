// Package config ...
package config

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
)

// Config - ...
type Config struct {
	ServerAddress     string `json:"server_address"`
	ServerAddressGRPC string `json:"server_address_grpc"`
	BaseURL           string `json:"base_url"`
	LogLevel          string `json:"log_level"`
	FileStoragePath   string `json:"file_storage_path"`
	DatabaseDsn       string `json:"database_dsn"`
	SecureConnection  bool   `json:"enable_https"`
	TrustedSubnet     string `json:"trusted_subnet"`
}

// C - ...
var C = Config{
	ServerAddress:     "localhost:8080",
	ServerAddressGRPC: "localhost:8081",
	BaseURL:           "http://localhost:8080/",
	LogLevel:          "info",
	FileStoragePath:   "",
	DatabaseDsn:       "",
	SecureConnection:  false,
	TrustedSubnet:     "0.0.0.0/0",
}

// Init - config initiator
func Init(ctx context.Context) {
	var config string

	def := C

	flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
	flagSet.StringVar(&config, "c", "", "name of config")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Println(err)
	}
	if b, err := os.ReadFile(config); err == nil {
		if err = json.Unmarshal(b, &C); err != nil {
			log.Println(err)
		}
	}

	d := C

	flag.StringVar(&config, "c", "", "name of config")
	flag.StringVar(&d.ServerAddress, "a", "localhost:8080", "Address of http server")
	flag.StringVar(&d.ServerAddressGRPC, "g", "localhost:8081", "Address of grpcHandler server")
	flag.StringVar(&d.BaseURL, "b", "http://localhost:8080/", "Response prefix")
	flag.StringVar(&d.LogLevel, "l", "info", "Set log level")
	flag.StringVar(&d.FileStoragePath, "f", "", "Storage file name")
	flag.StringVar(&d.DatabaseDsn, "d", "", "Database dsn")
	flag.BoolVar(&d.SecureConnection, "s", false, "")
	flag.StringVar(&d.TrustedSubnet, "t", "0.0.0.0/0", "Trusted subnet")

	flag.Parse()

	dv := reflect.ValueOf(d)
	Cv := reflect.ValueOf(&C).Elem()
	defV := reflect.ValueOf(def)
	for i := 0; i < dv.NumField(); i++ {
		if !dv.Field(i).Equal(defV.Field(i)) {
			Cv.Field(i).Set(dv.Field(i))
		}
	}

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		C.ServerAddress = v
	}
	if v, ok := os.LookupEnv("SERVER_ADDRESS_GRPC"); ok {
		C.ServerAddressGRPC = v
	}
	if v, ok := os.LookupEnv("BASE_URL"); ok {
		C.BaseURL = v
	}
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		C.LogLevel = v
	}
	if v, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		C.FileStoragePath = v
	}
	if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
		C.DatabaseDsn = v
	}
	if v, ok := os.LookupEnv("ENABLE_HTTPS"); ok && v == "YES" {
		C.SecureConnection = true
	}
	if v, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		C.TrustedSubnet = v
	}
}
