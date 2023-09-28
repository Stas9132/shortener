package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"io"
	"net/http"
	"net/url"
	"os"
	"shortener/config"
	"shortener/internal/app/model"
	"shortener/internal/logger"
	"sync"
)

var storage = sync.OnceValue(func() map[string][]byte {
	m := make(map[string][]byte)
	if f, e := os.Open(*config.FileStoragePath); e != nil {
		logger.Log.WithField("error", e).Errorln("Unable to read File Storage Path")
	} else {
		defer f.Close()
		var fStor model.FileStorageT
		if e = json.NewDecoder(f).Decode(&fStor); e != nil {
			logger.Log.WithField("error", e).Errorln("File storage is corrupted")
		} else {
			for _, r := range fStor {
				m[r.ShortURL] = []byte(r.OriginalURL)
			}
		}
	}
	return m
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
	h := getHash(b)
	storage()[h] = b
	u, _ := url.JoinPath(*config.BaseURL, h)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(u))
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
	h := getHash([]byte(request.URL.String()))
	storage()[h] = []byte(request.URL.String())
	response.Result, _ = url.JoinPath(*config.BaseURL, h)
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
	for short, orig := range storage() {
		lu = append(lu, model.ListURLRecordT{
			ShortURL: func() string {
				s, _ := url.JoinPath(*config.BaseURL, short)
				return s
			}(),
			OriginalURL: string(orig),
		})
	}

	if len(lu) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	fmt.Println(lu)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, lu)
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
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
