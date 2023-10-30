package storage

import (
	"context"
	"database/sql"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"shortener/config"
	"shortener/internal/logger"
)

type DBT struct {
	appCtx context.Context
	logger logger.Logger
	db     *sql.DB
}

func NewDB(ctx context.Context, l logger.Logger) *DBT {
	db, err := sql.Open("pgx", *config.DatabaseDsn)
	if err != nil {
		logger.WithField("error", err).Errorln("Error while open db")
	}
	_, _ = db.Exec("create table shortener(id serial primary key , short_url varchar(255) unique not null, original_url varchar(255))")
	return &DBT{
		appCtx: ctx,
		logger: l,
		db:     db,
	}
}

func (s *DBT) Load(key string) (value string, ok bool) {
	err := s.db.QueryRowContext(s.appCtx, "SELECT original_url FROM shortener WHERE short_url = $1", key).
		Scan(&value)
	if err != nil {
		return "", false
	}
	ok = true
	return
}

func (s *DBT) Store(key, value string) {
	_, err := s.db.ExecContext(s.appCtx, "INSERT INTO shortener(short_url,original_url) values ($1, $2)", key, value)

	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		s.logger.WithField("URL", value).Info("URL already exist")
	} else if err != nil {
		s.logger.WithField("error", err).
			Warningln("Error while insert data")
	}

}

func (s *DBT) LoadOrStore(key, value string) (actual string, loaded bool) {
	actual, loaded = s.Load(key)
	s.Store(key, value)
	return
}

func (s *DBT) Range(f func(key, value string) bool) {
	rows, err := s.db.QueryContext(s.appCtx, "SELECT short_url, original_url FROM shortener")
	if err != nil || rows.Err() != nil {
		s.logger.WithField("error", err).
			Warningln("Error while select data")
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		err = rows.Scan(&key, &value)
		if err != nil {
			s.logger.WithField("error", err).
				Warningln("Error while select data")
		}
		if !f(key, value) {
			break
		}
	}

}

func (s *DBT) Close() error {
	return s.db.Close()
}

func (s *DBT) Ping() error {
	return s.db.Ping()
}
