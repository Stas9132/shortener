package config

import "flag"

var (
	ServerAddress  *string
	ResponsePrefix *string
)

func InitConfig() {
	ServerAddress = flag.String("a", "localhost:8080", "Address of http server")
	ResponsePrefix = flag.String("b", "http://localhost:8080/", "Response prefix")
	flag.Parse()
}
