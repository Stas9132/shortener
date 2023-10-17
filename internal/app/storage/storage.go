package storage

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"shortener/config"
	"shortener/internal/logger"
)

type StorageT struct {
	cache map[any]any
	file  *os.File
	db    *sql.DB
}

type StorageI interface {
	Load(key any) (value any, ok bool)
	Store(key, value any)
	Range(f func(key, value any) bool)
	Ping() error
	Close() error
}

func New() *StorageT {
	c := make(map[any]any)

	f, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.WithField("error", err).Errorln("Error while open file")
	} else {
		var fd []FileStorageRecordT
		if err = json.NewDecoder(f).Decode(&fd); err != nil {
			logger.WithField("error", err).Errorln("Error while decode json")
		} else {
			for _, record := range fd {
				c[record.ShortURL] = record.OriginalURL
			}
		}
	}

	db, err := sql.Open("pgx", *config.DatabaseDsn)

	return &StorageT{
		cache: c,
		file:  f,
		db:    db,
	}
}

func (s *StorageT) Load(key any) (value any, ok bool) {
	v, ok := s.cache[key]
	return v, ok
}

func (s *StorageT) Store(key, value any) {
	s.cache[key] = value
	_, _ = s.file.Seek(0, 0)
	var fd []FileStorageRecordT
	if err := json.NewDecoder(s.file).Decode(&fd); err != nil {
		logger.WithField("error", err).Errorln("Error while decode json")
	}
	fd = append(fd, FileStorageRecordT{
		UUID:        uuid.NewString(),
		ShortURL:    key.(string),
		OriginalURL: value.(string),
	})
	_, _ = s.file.Seek(0, 0)
	if err := json.NewEncoder(s.file).Encode(fd); err != nil {
		logger.WithField("error", err).Errorln("Error while encode json")
	}
}

func (s *StorageT) Range(f func(key, value any) bool) {
	for k, v := range s.cache {
		if !f(k, v) {
			break
		}
	}
}

func (s *StorageT) Close() error {
	return s.file.Close()
}

func (s *StorageT) Ping() error {
	return s.db.Ping()
}

type FileStorageRecordT struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
