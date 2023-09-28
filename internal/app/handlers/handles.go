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
	"shortener/internal/app/model"
	"shortener/internal/logger"
	"sync"
)

var storage = sync.OnceValue(func() *model.FileStorageT {
	return &model.FileStorageT{}
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

func PostRoot(w http.ResponseWriter, r *http.Request) {
	b, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	shortURL, e := url.JoinPath(
		*config.BaseURL,
		getHash(b))
	if e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	storage().Add(shortURL, string(b))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func PostApiShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.WithField("method", r.Method).Infoln("got request with bad method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request model.Request
	var response model.Response

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL, err := url.JoinPath(
		*config.BaseURL,
		getHash([]byte(request.URL.String())))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	storage().Add(shortURL, request.URL.String())
	response.Result = shortURL
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func GetApiUserURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Log.WithField("method", r.Method).Infoln("got request with bad method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var lu model.ListURLs
	lr, err := storage().ListRecords()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, r := range lr {
		lu = append(lu, model.ListURLRecordT{
			ShortURL:    r.ShortURL,
			OriginalURL: r.OriginalURL,
		})
	}

	if len(lu) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, lu)
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
	f := chi.URLParam(r, "sn")
	s, err := storage().Get(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", s.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(s.OriginalURL))
}
