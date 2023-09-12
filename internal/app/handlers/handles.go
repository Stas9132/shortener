package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"io"
	"net/http"
	"net/url"
	"shortener/config"
	"sync"
)

var storage = sync.OnceValue(func() map[string][]byte {
	return make(map[string][]byte)
})

func getHash(b []byte) string {
	h := md5.Sum(b)
	d := make([]byte, len(h)/4)
	for i := range d {
		d[i] = h[i] + h[i+len(h)/4] + h[i+len(h)/2] + h[i+3*len(h)/4]
	}
	return hex.EncodeToString(d)
}

func Default(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	b, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	h := getHash(b)
	storage()[h] = b
	u, _ := url.JoinPath(*config.BaseURL, h)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(u))
}

func JSONHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL string `json:"url"`
	}
	var response struct {
		Result string `json:"result"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h := getHash([]byte(request.URL))
	storage()[h] = []byte(request.URL)
	response.Result, _ = url.JoinPath(*config.BaseURL, h)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func GetByShortName(w http.ResponseWriter, r *http.Request) {
	f := chi.URLParam(r, "sn")
	b, ok := storage()[f]
	if !ok {
		http.Error(w, f, http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", string(b))
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write(b)
}
