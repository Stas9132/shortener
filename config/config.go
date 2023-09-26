package config

import (
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
	fileStoragePath = "/tmp/short-url-db.json"
	FileStoragePath = &fileStoragePath
)

func InitConfig() {
	ServerAddress = flag.String("a", "localhost:8080", "Address of http server")
	BaseURL = flag.String("b", "http://localhost:8080/", "Response prefix")
	LogLevel = flag.String("l", "info", "Set log level")
	FileStoragePath = flag.String("f", "/tmp/short-url-db.json", "Storage file name")

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
}
