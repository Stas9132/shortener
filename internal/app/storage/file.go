package storage

import (
	"context"
	"encoding/json"
	"github.com/Stas9132/shortener/config"
	"github.com/Stas9132/shortener/internal/logger"
	"os"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// FileStorageT - struct
type FileStorageT struct {
	appCtx context.Context
	logger logger.Logger
	cache  map[string]FileStorageRecordT
	file   *os.File
}

// NewFileStorage - constructor
func NewFileStorage(ctx context.Context, l logger.Logger) (*FileStorageT, error) {
	c := make(map[string]FileStorageRecordT)
	var f *os.File

	if config.C.FileStoragePath != "" {
		var err error
		f, err = os.OpenFile(config.C.FileStoragePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil && len(config.C.FileStoragePath) > 0 {
			logger.WithField("error", err).Errorln("Error while open file")
			return nil, err
		}
		var fd []FileStorageRecordT
		if err = json.NewDecoder(f).Decode(&fd); err != nil && err.Error() != "EOF" && f != nil {
			logger.WithField("error", err).Errorln("Error while unmarshal json")
			return nil, err
		}
		for _, record := range fd {
			c[record.ShortURL] = record
		}
	}
	return &FileStorageT{
		appCtx: ctx,
		logger: l,
		cache:  c,
		file:   f,
	}, nil
}

// Load - method
func (s *FileStorageT) Load(key string) (string, bool) {
	value, ok := s.cache[key]
	return value.OriginalURL, ok
}

// Store - method
func (s *FileStorageT) Store(key, value string) {
	s.StoreExt(key, value, uuid.NewString())
}

// StoreExt - method
func (s *FileStorageT) StoreExt(key, value, user string) {
	s.cache[key] = FileStorageRecordT{OriginalURL: value, UUID: user}
	if s.file != nil {
		if _, err := s.file.Seek(0, 0); err != nil {
			s.logger.WithField("error", err).Errorln("Error while seek file")
		}
		var fd []FileStorageRecordT
		if err := json.NewDecoder(s.file).Decode(&fd); err != nil {
			s.logger.WithField("error", err).Errorln("Error while unmarshal json")
		}
		fd = append(fd, FileStorageRecordT{
			UUID:        user,
			ShortURL:    key,
			OriginalURL: value,
		})
		if _, err := s.file.Seek(0, 0); err != nil {
			s.logger.WithField("error", err).Errorln("Error while seek file")
		}
		if err := json.NewEncoder(s.file).Encode(fd); err != nil {
			s.logger.WithField("error", err).Errorln("Error while marshal json")
		}
	}
}

// LoadOrStore - method
func (s *FileStorageT) LoadOrStore(key, value string) (actual string, loaded bool) {
	actual, loaded = s.Load(key)
	s.Store(key, value)
	return
}

// LoadOrStoreExt - method
func (s *FileStorageT) LoadOrStoreExt(key, value, user string) (actual string, loaded bool) {
	actual, loaded = s.Load(key)
	s.StoreExt(key, value, user)
	return
}

// RangeExt - method
func (s *FileStorageT) RangeExt(f func(key, value, user string) bool) {
	for k, v := range s.cache {
		if !f(k, v.OriginalURL, v.UUID) {
			break
		}
	}
}

// Range - method
func (s *FileStorageT) Range(f func(key, value string) bool) {
	for k, v := range s.cache {
		if !f(k, v.OriginalURL) {
			break
		}
	}
}

// Close - method
func (s *FileStorageT) Close() error {
	return s.file.Close()
}

// Ping - method
func (s *FileStorageT) Ping() error {
	return nil
}

// Delete - method
func (s *FileStorageT) Delete(keys ...string) {
	for _, key := range keys {
		delete(s.cache, key)
		if _, err := s.file.Seek(0, 0); err != nil {
			s.logger.WithField("error", err).Errorln("Error while seek file")
		}
		var tfd, fd []FileStorageRecordT
		if err := json.NewDecoder(s.file).Decode(&fd); err != nil {
			s.logger.WithField("error", err).Errorln("Error while unmarshal json")
		}

		for _, t := range fd {
			if t.ShortURL != key {
				tfd = append(tfd, t)
			}
		}
		fd = tfd
		if _, err := s.file.Seek(0, 0); err != nil {
			s.logger.WithField("error", err).Errorln("Error while seek file")
		}
		if err := json.NewEncoder(s.file).Encode(fd); err != nil {
			s.logger.WithField("error", err).Errorln("Error while marshal json")
		}
	}
}

// FileStorageRecordT - type
type FileStorageRecordT struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
