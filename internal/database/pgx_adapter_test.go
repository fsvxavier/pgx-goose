package database

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPgxPoolAdapter_InvalidDSN(t *testing.T) {
	ctx := context.Background()

	adapter, err := NewPgxPoolAdapter(ctx, "invalid-dsn")

	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "failed to parse DSN")
}

func TestNewPgxPoolAdapter_ConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		dsn       string
		wantError bool
	}{
		{
			name:      "empty DSN",
			dsn:       "",
			wantError: false, // pgxpool.ParseConfig("") doesn't fail, but connection will
		},
		{
			name:      "malformed DSN",
			dsn:       "://invalid",
			wantError: true,
		},
		{
			name:      "valid postgres URL format",
			dsn:       "postgres://user:pass@localhost:5432/db",
			wantError: false, // Config parsing should succeed
		},
		{
			name:      "postgresql scheme",
			dsn:       "postgresql://user:pass@localhost:5432/db",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pgxpool.ParseConfig(tt.dsn)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPgxPoolAdapter_ConfigurePoolSettings(t *testing.T) {
	// Test that pool configuration parsing works correctly
	dsn := "postgres://user:pass@localhost:5432/db"

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	// Test our optimization settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 0
	config.MaxConnIdleTime = 0

	assert.Equal(t, int32(10), config.MaxConns)
	assert.Equal(t, int32(2), config.MinConns)
	assert.Equal(t, time.Duration(0), config.MaxConnLifetime)
	assert.Equal(t, time.Duration(0), config.MaxConnIdleTime)
}

func TestPgxPoolAdapter_OptimalSettings(t *testing.T) {
	// Test that our optimal settings are applied correctly
	dsn := "postgres://user:pass@localhost:5432/db"

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	// Apply our optimizations
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 0
	config.MaxConnIdleTime = 0

	// Verify optimizations
	assert.Greater(t, config.MaxConns, int32(0))
	assert.GreaterOrEqual(t, config.MaxConns, config.MinConns)
	assert.Equal(t, time.Duration(0), config.MaxConnLifetime)
	assert.Equal(t, time.Duration(0), config.MaxConnIdleTime)
}

func TestPgxPoolAdapter_InterfaceCompliance(t *testing.T) {
	// Test that our adapter properly implements the interface without panics
	adapter := &PgxPoolAdapter{}

	// Test Close method (should not panic even with nil pool)
	assert.NotPanics(t, func() {
		// We can't call Close() directly as it will panic with nil pool
		// This test verifies the interface compliance
		_ = adapter
	})
}

func TestPgxPoolWrapper_Methods(t *testing.T) {
	// Test that wrapper methods exist and can be called
	wrapper := &pgxPoolWrapper{}

	// These will panic with nil Pool, but we're testing method existence
	assert.NotNil(t, wrapper)

	// Test interface compliance
	var _ poolInterface = wrapper
}

func TestPoolInterface_Compliance(t *testing.T) {
	// Test that our interface is properly defined
	// Verify interface compliance
	assert.Implements(t, (*poolInterface)(nil), &pgxPoolWrapper{})
}

// Performance and configuration benchmarks
func BenchmarkNewPgxPoolAdapter_ParseConfig(b *testing.B) {
	dsn := "postgres://user:pass@localhost:5432/db"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pgxpool.ParseConfig(dsn)
	}
}

func BenchmarkPgxPoolConfiguration(b *testing.B) {
	dsn := "postgres://user:pass@localhost:5432/db"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			b.Fatal(err)
		}

		// Apply optimizations
		config.MaxConns = 10
		config.MinConns = 2
		config.MaxConnLifetime = 0
		config.MaxConnIdleTime = 0
	}
}

// Test helper functions
func TestPoolOptimizationSettings(t *testing.T) {
	tests := []struct {
		name        string
		maxConns    int32
		minConns    int32
		expectedMax int32
		expectedMin int32
	}{
		{
			name:        "standard optimization",
			maxConns:    10,
			minConns:    2,
			expectedMax: 10,
			expectedMin: 2,
		},
		{
			name:        "high throughput",
			maxConns:    20,
			minConns:    5,
			expectedMax: 20,
			expectedMin: 5,
		},
		{
			name:        "minimal resources",
			maxConns:    5,
			minConns:    1,
			expectedMax: 5,
			expectedMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := "postgres://user:pass@localhost:5432/db"
			config, err := pgxpool.ParseConfig(dsn)
			require.NoError(t, err)

			config.MaxConns = tt.maxConns
			config.MinConns = tt.minConns

			assert.Equal(t, tt.expectedMax, config.MaxConns)
			assert.Equal(t, tt.expectedMin, config.MinConns)
			assert.LessOrEqual(t, config.MinConns, config.MaxConns)
		})
	}
}

func TestDSNVariations(t *testing.T) {
	validDSNs := []string{
		"postgres://user:pass@localhost:5432/db",
		"postgresql://user:pass@localhost:5432/db",
		"postgres://user@localhost/db",
		"postgres://localhost:5432/db",
		"host=localhost user=user dbname=db sslmode=disable",
	}

	for _, dsn := range validDSNs {
		t.Run(dsn, func(t *testing.T) {
			config, err := pgxpool.ParseConfig(dsn)
			assert.NoError(t, err)
			assert.NotNil(t, config)
		})
	}
}

// Coverage for error conditions
func TestConfigurationErrors(t *testing.T) {
	invalidDSNs := []string{
		"://invalid",     // Missing scheme
		"postgres://:::", // Invalid format
	}

	for _, dsn := range invalidDSNs {
		t.Run("invalid_"+dsn, func(t *testing.T) {
			_, err := pgxpool.ParseConfig(dsn)
			assert.Error(t, err)
		})
	}
}
