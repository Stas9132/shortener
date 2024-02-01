// Package config ...
package config

import (
	"context"
	"flag"
	"os"
)

// ServerAddress - address of server
// BaseURL - base URL
// LogLevel - logging level
// FileStoragePath - path of file storage
// DatabaseDsn - data set name of databse

type Config struct {
	ServerAddress    string
	BaseURL          string
	LogLevel         string
	FileStoragePath  string
	DatabaseDsn      string
	SecureConnection bool
}

var C = Config{
	ServerAddress:    "localhost:8080",
	BaseURL:          "http://localhost:8080/",
	LogLevel:         "info",
	FileStoragePath:  "",
	DatabaseDsn:      "",
	SecureConnection: false,
}

// Init - config initiator
func Init(ctx context.Context) {
	flag.StringVar(&C.ServerAddress, "a", "localhost:8080", "Address of http server")
	flag.StringVar(&C.BaseURL, "b", "http://localhost:8080/", "Response prefix")
	flag.StringVar(&C.LogLevel, "l", "info", "Set log level")
	flag.StringVar(&C.FileStoragePath, "f", "", "Storage file name")
	flag.StringVar(&C.DatabaseDsn, "d", "", "Database dsn")
	flag.BoolVar(&C.SecureConnection, "s", false, "")

	flag.Parse()

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		C.ServerAddress = v
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
}
