package main

import (
	"net/http"
	"shortener/internal/app/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.MainHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
