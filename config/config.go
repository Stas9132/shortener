package config

import (
	"flag"
	"os"
)

var (
	serverAddress = "localhost:8080"
	ServerAddress = &serverAddress
	baseURL       = "http://localhost:8080/"
	BaseURL       = &baseURL
)

func InitConfig() {
	ServerAddress = flag.String("a", "localhost:8080", "Address of http server")
	BaseURL = flag.String("b", "http://localhost:8080/", "Response prefix")
	flag.Parse()

	if v := os.Getenv("SERVER_ADDRESS"); len(v) > 0 {
		*ServerAddress = v
	}
	if v := os.Getenv("BASE_URL"); len(v) > 0 {
		*BaseURL = v
	}
}
