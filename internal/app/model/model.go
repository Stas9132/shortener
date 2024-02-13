// Package model ...
package model

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/Stas9132/shortener/config"
	"github.com/Stas9132/shortener/internal/logger"
	"net/url"
	"strconv"
)

// Request struct
type Request struct {
	URL *url.URL `json:"url"`
}

// UnmarshalJSON - request method
func (r *Request) UnmarshalJSON(data []byte) (err error) {
	type RequestAlias Request

	aliasValue := &struct {
		*RequestAlias
		URL json.RawMessage `json:"url"`
	}{
		RequestAlias: (*RequestAlias)(r),
	}
	if err = json.Unmarshal(data, aliasValue); err != nil {
		return
	}
	u, err := strconv.Unquote(string(aliasValue.URL))
	if err != nil {
		return
	}
	r.URL, err = url.ParseRequestURI(u)
	return
}

// Response struct
type Response struct {
	Result string `json:"result"`
}

// ListURLs slice
type ListURLs []ListURLRecordT

// ListURLRecordT struct
type ListURLRecordT struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	User        string `json:"-"`
}

// Batch slice of struct
type Batch []struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url"`
}

// BatchDelete slice
type BatchDelete []string

// Stats struct
type Stats struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}

// Storage - ...
type Storage interface {
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

// ErrExist - ...
var ErrExist = errors.New("already exist")

// API - ...
type API struct {
	logger.Logger
	storage Storage
}

// NewAPI - ...
func NewAPI(logger logger.Logger, storage Storage) *API {
	return &API{Logger: logger, storage: storage}
}

// PostPlainText - api handler
func (a *API) PostPlainText(b []byte, issuer string) (string, error) {
	shortURL, e := url.JoinPath(
		config.C.BaseURL,
		getHash(b))
	if e != nil {
		a.WithFields(map[string]interface{}{
			"error": e,
		}).Warn("url.JoinPath error")
		return "", e
	}
	_, exist := a.storage.LoadOrStoreExt(shortURL, string(b), issuer)

	if exist {
		return shortURL, ErrExist
	}
	return shortURL, nil
}

func (a *API) Post(request Request, issuer string) (*Response, error) {

	shortURL, err := url.JoinPath(
		config.C.BaseURL,
		getHash([]byte(request.URL.String())))

	if err != nil {
		a.WithFields(map[string]interface{}{
			"error": err,
		}).Warn("url.JoinPath")
		return nil, err
	}

	_, exist := a.storage.LoadOrStoreExt(shortURL, request.URL.String(), issuer)

	response := &Response{}
	response.Result = shortURL
	if exist {
		return response, ErrExist
	}

	return response, nil
}

func (a *API) GetUserURLs(issuer string) (ListURLs, error) {
	var lu ListURLs
	a.storage.RangeExt(func(key, value, user string) bool {
		lu = append(lu, ListURLRecordT{
			ShortURL:    key,
			OriginalURL: value,
			User:        user,
		})
		return true
	})

	var tlu ListURLs
	for _, u := range lu {
		if u.User == issuer {
			tlu = append(tlu, u)
		}
	}
	lu = tlu

	return lu, nil
}

// getHash - ...
func getHash(b []byte) string {
	d := make([]byte, 4)
	for i, v := range b {
		d[i%4] += v
	}
	return hex.EncodeToString(d)
}
