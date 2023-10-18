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
	strg "shortener/internal/app/storage"
	"shortener/internal/logger"
)

type APII interface {
	Default(w http.ResponseWriter, r *http.Request)
	PostPlainText(w http.ResponseWriter, r *http.Request)
	PostJSON(w http.ResponseWriter, r *http.Request)
	GetUserURLs(w http.ResponseWriter, r *http.Request)
	GetRoot(w http.ResponseWriter, r *http.Request)
	GetPing(w http.ResponseWriter, r *http.Request)
	PostBatch(w http.ResponseWriter, r *http.Request)
}

type APIT struct {
	storage strg.StorageI
}

func NewAPI(storage strg.StorageI) APIT {
	return APIT{storage: storage}
}

func getHash(b []byte) string {
	h := md5.Sum(b)
	d := make([]byte, len(h)/4)
	for i := range d {
		d[i] = h[i] + h[i+len(h)/4] + h[i+len(h)/2] + h[i+3*len(h)/4]
	}
	return hex.EncodeToString(d)
}

func (a APIT) Default(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (a APIT) PostPlainText(w http.ResponseWriter, r *http.Request) {
	b, e := io.ReadAll(r.Body)
	if e != nil {
		logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": e}).
			Warn("io.ReadAll error")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	shortURL, e := url.JoinPath(
		*config.BaseURL,
		getHash(b))
	if e != nil {
		logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": e}).
			Warn("url.JoinPath error")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	a.storage.Store(shortURL, string(b))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (a APIT) PostJSON(w http.ResponseWriter, r *http.Request) {
	var request model.Request
	var response model.Response

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": err}).
			Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL, err := url.JoinPath(
		*config.BaseURL,
		getHash([]byte(request.URL.String())))
	if err != nil {
		logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": err}).
			Warn("url.JoinPath")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.storage.Store(shortURL, request.URL.String())
	response.Result = shortURL
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func (a APIT) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	var lu model.ListURLs
	a.storage.Range(func(key, value any) bool {
		lu = append(lu, model.ListURLRecordT{
			ShortURL:    key.(string),
			OriginalURL: value.(string),
		})
		return true
	})

	if len(lu) == 0 || r.Header.Get("Accept-Encoding") != "identity" {
		render.NoContent(w, r)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, lu)
}

func (a APIT) GetRoot(w http.ResponseWriter, r *http.Request) {
	shortURL, e := url.JoinPath(
		*config.BaseURL,
		chi.URLParam(r, "sn"))
	if e != nil {
		logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": e}).
			Warn("url.JoinPath")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	s, ok := a.storage.Load(shortURL)
	if !ok {
		render.NoContent(w, r)
		return
	}
	w.Header().Set("Location", s.(string))
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(s.(string)))
}

func (a APIT) GetPing(w http.ResponseWriter, r *http.Request) {
	err := a.storage.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a APIT) PostBatch(w http.ResponseWriter, r *http.Request) {
	var batch model.Batch

	err := json.NewDecoder(r.Body).Decode(&batch)
	if err != nil {
		logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": err}).
			Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i := range batch {
		batch[i].ShortURL, err = url.JoinPath(
			*config.BaseURL,
			getHash([]byte(batch[i].OriginalURL)))
		if err != nil {
			logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
				"uri":   r.RequestURI,
				"error": err}).
				Warn("url.JoinPath")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a.storage.Store(batch[i].ShortURL, batch[i].OriginalURL)
		batch[i].OriginalURL = ""
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, batch)
}
