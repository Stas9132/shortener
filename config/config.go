package config

import (
	"context"
	"flag"
	"os"
)

var (
	serverAddress   = "localhost:8080"
	ServerAddress   = &serverAddress
	baseURL         = "http://localhost:8080/"
	BaseURL         = &baseURL
	logLevel        = "info"
	LogLevel        = &logLevel
	fileStoragePath = ""
	FileStoragePath = &fileStoragePath
	databaseDsn     = ""
	DatabaseDsn     = &databaseDsn
)

func Init(ctx context.Context) {
	ServerAddress = flag.String("a", "localhost:8080", "Address of http server")
	BaseURL = flag.String("b", "http://localhost:8080/", "Response prefix")
	LogLevel = flag.String("l", "info", "Set log level")
	FileStoragePath = flag.String("f", "", "Storage file name")
	DatabaseDsn = flag.String("d", "", "Database dsn")

	flag.Parse()

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		ServerAddress = &v
	}
	if v, ok := os.LookupEnv("BASE_URL"); ok {
		BaseURL = &v
	}
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		LogLevel = &v
	}
	if v, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		FileStoragePath = &v
	}
	if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
		DatabaseDsn = &v
	}
}
