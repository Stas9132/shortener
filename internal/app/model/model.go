package model

import (
	"encoding/json"
	"net/url"
	"strconv"
)

type Request struct {
	URL *url.URL `json:"url"`
}

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

type Response struct {
	Result string `json:"result"`
}

type ListURLs []ListURLRecordT

type ListURLRecordT struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorageT []struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
