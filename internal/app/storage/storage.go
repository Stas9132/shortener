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
	file  struct{}
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
	var d *sql.DB

	switch {
	case len(*config.DatabaseDsn) == 0:
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
	case len(*config.DatabaseDsn) > 0:
		db, err := sql.Open("pgx", *config.DatabaseDsn)
		if err != nil {
			logger.WithField("error", err).Errorln("Error while open db")
		}
		d = db
		_, _ = db.Exec("create table shortener(id serial primary key , short_url varchar(255) unique not null, original_url varchar(255))")
	}
	return &StorageT{
		cache: c,
		file:  struct{}{},
		db:    d,
	}
}

func (s *StorageT) Load(key any) (value any, ok bool) {
	switch {
	case len(*config.DatabaseDsn) == 0:
		value, ok = s.cache[key]
	case len(*config.DatabaseDsn) > 0:
		err := s.db.QueryRow("SELECT original_url FROM shortener WHERE short_url = $1", key.(string)).
			Scan(&value)
		if err != nil {
			return "", false
		}
		ok = true
	}
	return
}

func (s *StorageT) Store(key, value any) {
	switch {
	case len(*config.DatabaseDsn) == 0:
		s.cache[key] = value
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
	case len(*config.DatabaseDsn) > 0:
		_, err := s.db.Exec("INSERT INTO shortener(short_url,original_url) values ($1, $2)", key.(string), value.(string))
		if err != nil {
			logger.WithField("error", err).
				Warningln("Error while insert data")
		}
	}
}

func (s *StorageT) Range(f func(key, value any) bool) {
	switch {
	case len(*config.DatabaseDsn) == 0:
		for k, v := range s.cache {
			if !f(k, v) {
				break
			}
		}
	case len(*config.DatabaseDsn) > 0:
		rows, err := s.db.Query("SELECT short_url, original_url FROM shortener")
		if err != nil || rows.Err() != nil {
			logger.WithField("error", err).
				Warningln("Error while select data")
		}
		defer rows.Close()
		for rows.Next() {
			var key, value string
			err = rows.Scan(&key, &value)
			if err != nil {
				logger.WithField("error", err).
					Warningln("Error while select data")
			}
			if !f(key, value) {
				break
			}
		}

	}
}

func (s *StorageT) Close() error {
	return s.db.Close()
}

func (s *StorageT) Ping() error {
	return s.db.Ping()
}

type FileStorageRecordT struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
