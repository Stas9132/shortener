package storage

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/google/uuid"
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
		_ = m.Force(1)
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
	var b *bool
	err := s.db.QueryRowContext(s.appCtx, "SELECT original_url, is_deleted FROM shortener WHERE short_url = $1", key).
		Scan(&value, &b)
	if err != nil {
		s.logger.WithField("error", err).Errorln("error load()")
		return "", false
	}
	if b == nil || !*b {
		ok = true
	}
	return
}

func (s *DBT) StoreExt(key, value, user string) {
	_, err := s.db.ExecContext(s.appCtx, "INSERT INTO shortener(short_url,original_url, user_id) values ($1, $2, $3)", key, value, user)

	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		s.logger.WithField("URL", value).Info("URL already exist")
	} else if err != nil {
		s.logger.WithField("error", err).
			Warningln("Error while insert data")
	}
}

func (s *DBT) Store(key, value string) {
	s.StoreExt(key, value, uuid.NewString())
}

func (s *DBT) LoadOrStore(key, value string) (actual string, loaded bool) {
	actual, loaded = s.Load(key)
	s.Store(key, value)
	return
}

func (s *DBT) LoadOrStoreExt(key, value, user string) (actual string, loaded bool) {
	actual, loaded = s.Load(key)
	s.StoreExt(key, value, user)
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

func (s *DBT) RangeExt(f func(key, value, user string) bool) {
	rows, err := s.db.QueryContext(s.appCtx, "SELECT short_url, original_url, user_id FROM shortener")
	if err != nil || rows.Err() != nil {
		s.logger.WithField("error", err).
			Warningln("Error while select data")
	}
	defer rows.Close()
	for rows.Next() {
		var key, value, userId string
		err = rows.Scan(&key, &value, &userId)
		if err != nil {
			s.logger.WithField("error", err).
				Warningln("Error while select data")
		}
		if !f(key, value, userId) {
			break
		}
	}

}

func (s *DBT) Delete(keys ...string) {
	for _, key := range keys {
		_, err := s.db.ExecContext(s.appCtx, "update shortener set is_deleted = true where short_url = $1", key)
		if err != nil {
			s.logger.WithField("error", err).Errorln("error while db records mark as deleted")
		}
	}
}

func (s *DBT) Close() error {
	//return errors.Join(s.db.Close(), s.m.Down())
	return s.db.Close()
}

func (s *DBT) Ping() error {
	return s.db.Ping()
}
