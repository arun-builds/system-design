package database

import (
	"context"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns    = 50
	defaultMinConns    = 5
	defaultConnMaxLife = time.Hour
)

type DB struct {
	Write *pgxpool.Pool
	Reads []*pgxpool.Pool
}

func ConnectDB(masterDSN, replicaDSN string) (*DB, error) {
	ctx := context.Background()

	master, err := newPool(ctx, masterDSN)
	if err != nil {
		return nil, err
	}

	replica, err := newPool(ctx, replicaDSN)
	if err != nil {
		master.Close()
		return nil, err
	}

	return &DB{
		Write: master,
		Reads: []*pgxpool.Pool{master, replica},
	}, nil
}

func newPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = defaultMaxConns
	cfg.MinConns = defaultMinConns
	cfg.MaxConnLifetime = defaultConnMaxLife
	return pgxpool.NewWithConfig(ctx, cfg)
}

func (db *DB) Reader() *pgxpool.Pool {
	return db.Reads[rand.Intn(len(db.Reads))]
}

func (db *DB) Close() error {
	db.Write.Close()

	for _, r := range db.Reads {
		if r != db.Write {
			r.Close()
		}
	}

	return nil
}
