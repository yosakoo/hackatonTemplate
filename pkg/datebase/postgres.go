package database

import (
	"context"
	"fmt"
	"net/url"

	"hackathonTemplate/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgreSQL(ctx context.Context, poolConfig *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func NewPoolConfigFromDatabaseConfig(dbConfig config.DatabaseConfig) (*pgxpool.Config, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(dbConfig.User),
		url.QueryEscape(dbConfig.Password),
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database,
		dbConfig.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(dbConfig.MaxOpenConns)
	poolConfig.MinConns = int32(dbConfig.MaxIdleConns)
	poolConfig.MaxConnLifetime = dbConfig.MaxLifetime

	return poolConfig, nil
}

func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
