package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/Stas9132/shortener/config"
	"github.com/Stas9132/shortener/internal/app/handlers/middlware"
	"github.com/Stas9132/shortener/internal/app/model"
	"github.com/Stas9132/shortener/internal/logger"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
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
	logger  logger.Logger
}

// NewAPI() - constructor
func NewAPI(ctx context.Context, l logger.Logger, storage StorageI) APIT {
	return APIT{storage: storage, logger: l}
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
		a.logger.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      e,
		}).Warn("io.ReadAll error")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	shortURL, e := url.JoinPath(
		*config.BaseURL,
		getHash(b))
	if e != nil {
		a.logger.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      e,
		}).Warn("url.JoinPath error")
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	_, exist := a.storage.LoadOrStoreExt(shortURL, string(b), middlware.GetIssuer(r.Context()).ID)

	if exist {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(shortURL))
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}

// PostJSON - api handler
func (a APIT) PostJSON(w http.ResponseWriter, r *http.Request) {
	var request model.Request
	var response model.Response

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.logger.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      err,
		}).Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL, err := url.JoinPath(
		*config.BaseURL,
		getHash([]byte(request.URL.String())))
	if err != nil {
		a.logger.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      err,
		}).Warn("url.JoinPath")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, exist := a.storage.LoadOrStoreExt(shortURL, request.URL.String(), middlware.GetIssuer(r.Context()).ID)

	response.Result = shortURL
	if exist {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, response)
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

// GetUserURLs - api handler
func (a APIT) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	var lu model.ListURLs
	a.storage.RangeExt(func(key, value, user string) bool {
		lu = append(lu, model.ListURLRecordT{
			ShortURL:    key,
			OriginalURL: value,
			User:        user,
		})
		return true
	})

	switch middlware.GetIssuer(r.Context()).State {
	case "NEW":
		w.WriteHeader(http.StatusUnauthorized)
		return
	case "ESTABLISHED":
		var tlu model.ListURLs
		for _, u := range lu {
			if u.User == middlware.GetIssuer(r.Context()).ID {
				tlu = append(tlu, u)
			}
		}
		lu = tlu
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
		*config.BaseURL,
		chi.URLParam(r, "sn"))
	if e != nil {
		a.logger.WithFields(map[string]interface{}{
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
		a.logger.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      err,
		}).Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i := range batch {
		batch[i].ShortURL, err = url.JoinPath(
			*config.BaseURL,
			getHash([]byte(batch[i].OriginalURL)))
		if err != nil {
			a.logger.WithFields(map[string]interface{}{
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
		a.logger.WithFields(map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"uri":        r.RequestURI,
			"error":      err,
		}).Warn("json.Decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range batch {
		batch[i], err = url.JoinPath(*config.BaseURL, batch[i])
		if err != nil {
			a.logger.WithFields(map[string]interface{}{
				"remoteAddr": r.RemoteAddr,
				"uri":        r.RequestURI,
				"error":      err,
			}).Warn("url.JoinPath")
		}
	}

	go a.storage.Delete(batch...)

	w.WriteHeader(http.StatusAccepted)
}
