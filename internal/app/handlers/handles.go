package handlers

import (
	"context"
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

type StorageI interface {
	Load(key string) (value string, ok bool)
	Store(key, value string)
	Range(f func(key, value string) bool)
	LoadOrStore(key, value string) (actual string, loaded bool)
	Ping() error
	Close() error
}

type APIT struct {
	storage StorageI
	logger  logger.Logger
}

func NewAPI(ctx context.Context, l logger.Logger, storage StorageI) APIT {
	return APIT{storage: storage, logger: l}
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
		a.logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
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
		a.logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": e}).
			Warn("url.JoinPath error")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	_, exist := a.storage.LoadOrStore(shortURL, string(b))
	if exist {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(shortURL))
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

func (a APIT) PostJSON(w http.ResponseWriter, r *http.Request) {
	var request model.Request
	var response model.Response

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
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
		a.logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
			"uri":   r.RequestURI,
			"error": err}).
			Warn("url.JoinPath")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, exist := a.storage.LoadOrStore(shortURL, request.URL.String())

	response.Result = shortURL
	if exist {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, response)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func (a APIT) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	var lu model.ListURLs
	a.storage.Range(func(key, value string) bool {
		lu = append(lu, model.ListURLRecordT{
			ShortURL:    key,
			OriginalURL: value,
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
		a.logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
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
	w.Header().Set("Location", s)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(s))
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
		a.logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
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
			a.logger.WithFields(map[string]interface{}{"remoteAddr": r.RemoteAddr,
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
