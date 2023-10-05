package model

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/url"
	"os"
	"shortener/config"
	"strconv"
	"sync"
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

type FileStorageRecordT struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorageT struct {
	sync.Mutex
	records []FileStorageRecordT
}

func (f *FileStorageT) ListRecords() ([]FileStorageRecordT, error) {
	f.Lock()
	defer f.Unlock()
	file, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return f.records, json.NewDecoder(file).Decode(&f.records)
}

func (f *FileStorageT) Add(ShortURL, OriginalURL string) error {
	f.Lock()
	defer f.Unlock()
	f.records = append(f.records, FileStorageRecordT{
		UUID:        uuid.New().String(),
		ShortURL:    ShortURL,
		OriginalURL: OriginalURL,
	})
	file, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(f.records)
}

func (f *FileStorageT) Get(ShortURL string) (FileStorageRecordT, error) {
	f.Lock()
	defer f.Unlock()
	file, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return FileStorageRecordT{}, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&f.records)
	if err != nil {
		return FileStorageRecordT{}, err
	}
	for _, record := range f.records {
		if ShortURL == record.ShortURL {
			return record, nil
		}
	}
	return FileStorageRecordT{}, errors.New("not found")
}
