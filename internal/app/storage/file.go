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
	logger logger.Logger
	cache  map[any]any
	file   *os.File
}

func NewFileStorage(ctx context.Context, l logger.Logger) *FileStorageT {
	c := make(map[any]any)
	var f *os.File

	if len(*config.FileStoragePath) > 0 {
		var err error
		f, err = os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logger.WithField("error", err).Errorln("Error while open file")
		}
		var fd []FileStorageRecordT
		if err = json.NewDecoder(f).Decode(&fd); err != nil {
			logger.WithField("error", err).Errorln("Error while unmarshal json")
		}
		for _, record := range fd {
			c[record.ShortURL] = record.OriginalURL
		}
	}
	return &FileStorageT{
		appCtx: ctx,
		logger: l,
		cache:  c,
		file:   f,
	}
}

func (s *FileStorageT) Load(key any) (value any, ok bool) {
	value, ok = s.cache[key]
	return
}

func (s *FileStorageT) Store(key, value any) {
	s.cache[key] = value
	if s.file != nil {
		if _, err := s.file.Seek(0, 0); err != nil {
			s.logger.WithField("error", err).Errorln("Error while seek file")
		}
		var fd []FileStorageRecordT
		if err := json.NewDecoder(s.file).Decode(&fd); err != nil {
			s.logger.WithField("error", err).Errorln("Error while unmarshal json")
		}
		fd = append(fd, FileStorageRecordT{
			UUID:        uuid.NewString(),
			ShortURL:    key.(string),
			OriginalURL: value.(string),
		})
		if _, err := s.file.Seek(0, 0); err != nil {
			s.logger.WithField("error", err).Errorln("Error while seek file")
		}
		if err := json.NewEncoder(s.file).Encode(fd); err != nil {
			s.logger.WithField("error", err).Errorln("Error while marshal json")
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
	return s.file.Close()
}

func (s *FileStorageT) Ping() error {
	return nil
}

type FileStorageRecordT struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
