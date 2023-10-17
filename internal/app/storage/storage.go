package storage

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"os"
	"shortener/config"
	"sync"
)

type StorageT struct {
	cache map[any]any
}

type StorageI interface {
	Load(key any) (value any, ok bool)
	Store(key, value any)
	Range(f func(key, value any) bool)
	Close() error
}

func New() *StorageT {
	return &StorageT{cache: make(map[any]any)}
}

func (s *StorageT) Load(key any) (value any, ok bool) {
	v, ok := s.cache[key]
	return v, ok
}
func (s *StorageT) Store(key, value any) {
	s.cache[key] = value
}
func (s *StorageT) Range(f func(key, value any) bool) {
	for k, v := range s.cache {
		if !f(k, v) {
			break
		}
	}
}

func (s *StorageT) Close() error {
	return nil
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
