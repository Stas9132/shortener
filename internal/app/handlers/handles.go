// Package handlers ...
package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/Stas9132/shortener/config"
	"github.com/Stas9132/shortener/internal/app/handlers/middleware"
	"github.com/Stas9132/shortener/internal/app/model"
	"github.com/Stas9132/shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"io"
	"net/http"
	"net/url"
)

// APII main interface for handler
type APII interface {
	Default(w http.ResponseWriter, r *http.Request)
	PostPlainText(w http.ResponseWriter, r *http.Request)
	PostJSON(w http.ResponseWriter, r *http.Request)
	GetUserURLs(w http.ResponseWriter, r *http.Request)
	GetRoot(w http.ResponseWriter, r *http.Request)
	GetPing(w http.ResponseWriter, r *http.Request)
	PostBatch(w http.ResponseWriter, r *http.Request)
	DeleteUserUrls(w http.ResponseWriter, r *http.Request)
	GetStats(w http.ResponseWriter, r *http.Request)
}

type ModelAPI interface {
	PostPlainText(b []byte, issuer string) (string, error)
	Post(request model.Request, issuer string) (*model.Response, error)
	GetUserURLs(ctx context.Context) (model.ListURLs, error)
}

// StorageI - interface to storage
type StorageI interface {
	Load(key string) (value string, ok bool)
	Store(key, value string)
	RangeExt(f func(key, value, user string) bool)
	Range(f func(key, value string) bool)
	LoadOrStore(key, value string) (actual string, loaded bool)
	LoadOrStoreExt(key, value, user string) (actual string, loaded bool)
	Delete(keys ...string)
	Ping() error
	Close() error
}

// APIT - struct with api handlers
type APIT struct {
	storage StorageI
	logger.Logger
	m ModelAPI
}

// NewAPI() - constructor
func NewAPI(ctx context.Context, l logger.Logger, storage StorageI, model ModelAPI) APIT {
	return APIT{storage: storage, Logger: l, m: model}
}

//func getHash(b []byte) string {
//	h := md5.Sum(b)
//	d := make([]byte, len(h)/4)
//	for i := range d {
//		d[i] = h[i] + h[i+len(h)/4] + h[i+len(h)/2] + h[i+3*len(h)/4]
//	}
//	return hex.EncodeToString(d)
//}

func getHash(b []byte) string {
	d := make([]byte, 4)
	for i, v := range b {
		d[i%4] += v
	}
	return hex.EncodeToString(d)
}

// Default - api handler
func (a APIT) Default(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

// PostPlainText - api handler
func (a APIT) PostPlainText(w http.ResponseWriter, r *http.Request) {
	b, e := io.ReadAll(r.Body)
	if e != nil {
		a.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      e,
		}).Warn("io.ReadAll error")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	resp, e := a.m.PostPlainText(b, middleware.GetIssuer(r.Context()).ID)

	if e != nil {
		if !errors.Is(e, model.ErrExist) {
			a.WithFields(map[string]interface{}{
				"remoteAddr": r.RemoteAddr,
				"uri":        r.RequestURI,
				"error":      e,
			}).Warn("url.JoinPath error")
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(resp))
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(resp))
}

// PostJSON - api handler
func (a APIT) PostJSON(w http.ResponseWriter, r *http.Request) {
	var request model.Request

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      err,
		}).Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := a.m.Post(request, middleware.GetIssuer(r.Context()).ID)

	if err != nil {
		if !errors.Is(err, model.ErrExist) {
			a.WithFields(map[string]interface{}{
				"remoteAddr": r.RemoteAddr,
				"uri":        r.RequestURI,
				"error":      err,
			}).Warn("url.JoinPath error")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, response)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

// GetUserURLs - api handler
func (a APIT) GetUserURLs(w http.ResponseWriter, r *http.Request) {

	lu, err := a.m.GetUserURLs(r.Context())
	if err != nil {
		if errors.Is(err, model.ErrUnauthorized) {
			a.WithFields(map[string]interface{}{
				"remoteAddr": r.RemoteAddr,
				"uri":        r.RequestURI,
				"error":      err,
			}).Warn("model.GetUserURLs error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(lu) == 0 {
		render.NoContent(w, r)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, lu)
}

// GetRoot - api handler
func (a APIT) GetRoot(w http.ResponseWriter, r *http.Request) {
	shortURL, e := url.JoinPath(
		config.C.BaseURL,
		chi.URLParam(r, "sn"))
	if e != nil {
		a.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      e,
		}).Warn("url.JoinPath")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	s, ok := a.storage.Load(shortURL)
	if !ok {
		w.WriteHeader(http.StatusGone)
		return
	}
	w.Header().Set("Location", s)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(s))
}

// GetPing - api handler
func (a APIT) GetPing(w http.ResponseWriter, r *http.Request) {
	err := a.storage.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// PostBatch - api handler
func (a APIT) PostBatch(w http.ResponseWriter, r *http.Request) {
	var batch model.Batch

	err := json.NewDecoder(r.Body).Decode(&batch)
	if err != nil {
		a.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      err,
		}).Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i := range batch {
		batch[i].ShortURL, err = url.JoinPath(
			config.C.BaseURL,
			getHash([]byte(batch[i].OriginalURL)))
		if err != nil {
			a.WithFields(map[string]interface{}{
				"remoteAddr": r.RemoteAddr,
				"uri":        r.RequestURI,
				"error":      err,
			}).Warn("url.JoinPath")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a.storage.Store(batch[i].ShortURL, batch[i].OriginalURL)
		batch[i].OriginalURL = ""
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, batch)
}

// DeleteUserUrls - api handler
func (a APIT) DeleteUserUrls(w http.ResponseWriter, r *http.Request) {
	var batch model.BatchDelete

	err := json.NewDecoder(r.Body).Decode(&batch)
	if err != nil {
		a.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      err,
		}).Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range batch {
		batch[i], err = url.JoinPath(config.C.BaseURL, batch[i])
		if err != nil {
			a.WithFields(map[string]interface{}{
				"remoteAddr": r.RemoteAddr,
				"uri":        r.RequestURI,
				"error":      err,
			}).Warn("url.JoinPath")
		}
	}

	go a.storage.Delete(batch...)

	w.WriteHeader(http.StatusAccepted)
}

// GetStats - api handler
func (a APIT) GetStats(w http.ResponseWriter, r *http.Request) {
	users := make(map[string]struct{})
	urls := make(map[string]struct{})

	a.storage.RangeExt(func(key, value, user string) bool {
		users[user] = struct{}{}
		urls[value] = struct{}{}
		return true
	})

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(model.Stats{
		Urls:  len(urls),
		Users: len(users),
	}); err != nil {
		a.WithField(
			"error", err,
		).Warn("Unable encode json")
	}

}
