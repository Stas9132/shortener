package storage

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"shortener/config"
	"shortener/internal/logger"
)

type FileStorageT struct {
	appCtx context.Context
	cache  map[any]any
	file   struct {
		isPresent bool
	}
}

func NewFileStorage(ctx context.Context) *FileStorageT {
	c := make(map[any]any)
	var pf bool

	if len(*config.FileStoragePath) > 0 {
		b, err := os.ReadFile(*config.FileStoragePath)
		if err != nil {
			logger.WithField("error", err).Errorln("Error while read file")
		}
		var fd []FileStorageRecordT
		if err = json.Unmarshal(b, &fd); err != nil {
			logger.WithField("error", err).Errorln("Error while unmarshal json")
		}
		for _, record := range fd {
			c[record.ShortURL] = record.OriginalURL
		}
		pf = true
	}
	return &FileStorageT{
		appCtx: ctx,
		cache:  c,
		file:   struct{ isPresent bool }{isPresent: pf},
	}
}

func (s *FileStorageT) Load(key any) (value any, ok bool) {
	value, ok = s.cache[key]
	return
}

func (s *FileStorageT) Store(key, value any) {
	s.cache[key] = value
	if s.file.isPresent {
		b, err := os.ReadFile(*config.FileStoragePath)
		if err != nil {
			logger.WithField("error", err).Errorln("Error while read file")
		}
		var fd []FileStorageRecordT
		if err = json.Unmarshal(b, &fd); err != nil {
			logger.WithField("error", err).Errorln("Error while unmarshal json")
		}
		fd = append(fd, FileStorageRecordT{
			UUID:        uuid.NewString(),
			ShortURL:    key.(string),
			OriginalURL: value.(string),
		})
		if b, err = json.Marshal(fd); err != nil {
			logger.WithField("error", err).Errorln("Error while marshal json")
		}
		if err = os.WriteFile(*config.FileStoragePath, b, 0644); err != nil {
			logger.WithField("error", err).Errorln("Error while write file")
		}
	}
}

func (s *FileStorageT) LoadOrStore(key, value any) (actual any, loaded bool) {
	actual, loaded = s.Load(key)
	s.Store(key, value)
	return
}

func (s *FileStorageT) Range(f func(key, value any) bool) {
	for k, v := range s.cache {
		if !f(k, v) {
			break
		}
	}
}

func (s *FileStorageT) Close() error {
	return nil
}

func (s *FileStorageT) Ping() error {
	return nil
}

type FileStorageRecordT struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
