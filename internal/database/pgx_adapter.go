package database

import (
	"context"
	"fmt"

	"github.com/fsvxavier/pgx-goose/internal/interfaces"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxPoolAdapter adapts pgxpool.Pool to our DatabasePool interface.
type PgxPoolAdapter struct {
	pool poolInterface
}

// poolInterface allows for testing with mocks.
type poolInterface interface {
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close()
	Stat() *pgxpool.Stat
}

// NewPgxPoolAdapter creates a new PGX pool adapter.
func NewPgxPoolAdapter(ctx context.Context, dsn string) (interfaces.DatabasePool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Optimize pool settings for code generation workload
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 0
	config.MaxConnIdleTime = 0

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &PgxPoolAdapter{pool: &pgxPoolWrapper{pool}}, nil
}

func (p *PgxPoolAdapter) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *PgxPoolAdapter) Query(ctx context.Context, sql string, args ...interface{}) (interfaces.QueryResult, error) {
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &PgxRowsAdapter{rows: rows}, nil
}

func (p *PgxPoolAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) interfaces.Row {
	row := p.pool.QueryRow(ctx, sql, args...)
	return &PgxRowAdapter{row: row}
}

func (p *PgxPoolAdapter) Close() {
	p.pool.Close()
}

func (p *PgxPoolAdapter) Stats() interfaces.PoolStats {
	stats := p.pool.Stat()
	return interfaces.PoolStats{
		AcquireCount:         stats.AcquireCount(),
		AcquireDuration:      float64(stats.AcquireDuration().Nanoseconds()) / 1e6,
		AcquiredConns:        stats.AcquiredConns(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:    stats.ConstructingConns(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		IdleConns:            stats.IdleConns(),
		MaxConns:             stats.MaxConns(),
		TotalConns:           stats.TotalConns(),
	}
}

// pgxPoolWrapper wraps *pgxpool.Pool to implement poolInterface.
type pgxPoolWrapper struct {
	*pgxpool.Pool
}

// PgxRowsAdapter adapts pgx.Rows to our QueryResult interface.
type PgxRowsAdapter struct {
	rows pgx.Rows
}

func (r *PgxRowsAdapter) Next() bool {
	return r.rows.Next()
}

func (r *PgxRowsAdapter) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *PgxRowsAdapter) Close() {
	r.rows.Close()
}

func (r *PgxRowsAdapter) Err() error {
	return r.rows.Err()
}

// PgxRowAdapter adapts pgx.Row to our Row interface.
type PgxRowAdapter struct {
	row pgx.Row
}

func (r *PgxRowAdapter) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}
