package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"shortener/config"
)

type StorageT struct {
	appCtx context.Context
	cache  map[any]any
	file   struct{}
	db     *sql.DB
}

type StorageI interface {
	Load(key any) (value any, ok bool)
	Store(key, value any)
	Range(f func(key, value any) bool)
	LoadOrStore(key, value any) (actual any, loaded bool)
	Ping() error
	Close() error
}

func New(ctx context.Context) StorageI {
	if len(*config.DatabaseDsn) == 0 {
		return NewFileStorage(ctx)
	}
	return NewDB(ctx)
}
