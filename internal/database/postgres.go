package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidConfiguration = errors.New("invalid PostgreSQL configuration")
	ErrNotInitialized       = errors.New("PostgreSQL adapter is not initialized")
	ErrProbeFailed          = errors.New("PostgreSQL probe failed")
)

type PostgreSQL struct {
	pool *pgxpool.Pool
}

func NewPostgreSQL(ctx context.Context, connectionString string) (*PostgreSQL, error) {
	if strings.TrimSpace(connectionString) == "" {
		return nil, ErrInvalidConfiguration
	}

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}

	config.MinConns = 0
	config.MaxConns = 8
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}

	return &PostgreSQL{pool: pool}, nil
}

func (p *PostgreSQL) Probe(ctx context.Context) error {
	if p == nil || p.pool == nil {
		return ErrNotInitialized
	}

	probeContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := p.pool.Ping(probeContext); err != nil {
		return ErrProbeFailed
	}
	return nil
}

func (p *PostgreSQL) Close() {
	if p == nil || p.pool == nil {
		return
	}
	p.pool.Close()
}
