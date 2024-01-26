// Package model ...
package model

import (
	"encoding/json"
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
