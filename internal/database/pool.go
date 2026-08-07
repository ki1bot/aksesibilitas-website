package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(
	ctx context.Context,
	databaseURL string,
) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"konfigurasi PostgreSQL tidak valid: %w",
			err,
		)
	}

	config.MaxConns = 4
	config.MinConns = 0
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal membuat pool PostgreSQL: %w",
			err,
		)
	}

	pingContext, cancel := context.WithTimeout(
		ctx,
		5*time.Second,
	)
	defer cancel()

	if err := pool.Ping(pingContext); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"PostgreSQL tidak dapat dihubungi: %w",
			err,
		)
	}

	return pool, nil
}
