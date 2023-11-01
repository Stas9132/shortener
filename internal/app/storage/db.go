package storage

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"shortener/config"
	"shortener/internal/logger"
)

type DBT struct {
	appCtx context.Context
	logger logger.Logger
	db     *sql.DB
	m      *migrate.Migrate
}

func NewDB(ctx context.Context, l logger.Logger) *DBT {
	db, err := sql.Open("pgx", *config.DatabaseDsn)
	if err != nil {
		logger.WithField("error", err).Errorln("Error while open db")
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.WithField("error", err).Errorln("Error while get driver")
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/app/storage/migration",
		"pgx://"+*config.DatabaseDsn, driver)
	if err != nil {
		logger.WithField("error", err).Errorln("Error while create migrate")
	} else {
		if err = m.Up(); err != nil {
			logger.WithField("error", err).Errorln("Error while migrate up")
		}
	}

	return &DBT{
		appCtx: ctx,
		logger: l,
		db:     db,
		m:      m,
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

func (s *DBT) LoadOrStoreExt(key, value, user string) (actual string, loaded bool) {
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

func (s *DBT) RangeExt(f func(key, value, user string) bool) {}

func (s *DBT) Close() error {
	//return errors.Join(s.db.Close(), s.m.Down())
	return s.db.Close()
}

func (s *DBT) Ping() error {
	return s.db.Ping()
}
