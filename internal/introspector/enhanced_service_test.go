package introspector

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIntrospectorService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	config := ServiceConfig{
		Pool:   nil, // Will use mock in real tests
		Schema: "test_schema",
		Logger: logger,
	}

	service := NewIntrospectorService(config)
	require.NotNil(t, service)
	assert.Equal(t, "test_schema", service.schema)
	assert.NotNil(t, service.logger)
}

func TestNewIntrospectorService_DefaultSchema(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	config := ServiceConfig{
		Pool:   nil,
		Schema: "", // Empty schema should default to "public"
		Logger: logger,
	}

	service := NewIntrospectorService(config)
	require.NotNil(t, service)
	assert.Equal(t, "public", service.schema)
}

func TestMapPostgresToGoTypeEnhanced(t *testing.T) {
	tests := []struct {
		postgresType string
		isNullable   bool
		expected     string
	}{
		{"integer", false, "int32"},
		{"integer", true, "*int32"},
		{"bigint", false, "int64"},
		{"bigint", true, "*int64"},
		{"smallint", false, "int16"},
		{"smallint", true, "*int16"},
		{"text", false, "string"},
		{"text", true, "*string"},
		{"varchar", false, "string"},
		{"varchar", true, "*string"},
		{"character varying", false, "string"},
		{"character varying", true, "*string"},
		{"boolean", false, "bool"},
		{"boolean", true, "*bool"},
		{"bool", false, "bool"},
		{"bool", true, "*bool"},
		{"timestamp", false, "time.Time"},
		{"timestamp", true, "*time.Time"},
		{"timestamp without time zone", false, "time.Time"},
		{"timestamp without time zone", true, "*time.Time"},
		{"timestamp with time zone", false, "time.Time"},
		{"timestamp with time zone", true, "*time.Time"},
		{"timestamptz", false, "time.Time"},
		{"timestamptz", true, "*time.Time"},
		{"date", false, "time.Time"},
		{"date", true, "*time.Time"},
		{"uuid", false, "string"},
		{"uuid", true, "*string"},
		{"json", false, "json.RawMessage"},
		{"json", true, "json.RawMessage"},
		{"jsonb", false, "json.RawMessage"},
		{"jsonb", true, "json.RawMessage"},
		{"decimal", false, "float64"},
		{"decimal", true, "*float64"},
		{"numeric", false, "float64"},
		{"numeric", true, "*float64"},
		{"real", false, "float32"},
		{"real", true, "*float32"},
		{"float4", false, "float32"},
		{"float4", true, "*float32"},
		{"double precision", false, "float64"},
		{"double precision", true, "*float64"},
		{"float8", false, "float64"},
		{"float8", true, "*float64"},
		{"bytea", false, "[]byte"},
		{"bytea", true, "[]byte"},
		{"unknown_type", false, "interface{}"},
		{"unknown_type", true, "interface{}"},
	}

	for _, tt := range tests {
		nullable := "false"
		if tt.isNullable {
			nullable = "true"
		}
		t.Run(tt.postgresType+"_nullable_"+nullable, func(t *testing.T) {
			result := mapPostgresToGoType(tt.postgresType, tt.isNullable)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapPostgresToGoType_EdgeCases(t *testing.T) {
	// Test case sensitivity - our function is case sensitive
	assert.Equal(t, "interface{}", mapPostgresToGoType("INTEGER", false))
	assert.Equal(t, "interface{}", mapPostgresToGoType("Int4", false))

	// Test with lowercase
	assert.Equal(t, "string", mapPostgresToGoType("text", false))

	// Test unknown types
	assert.Equal(t, "interface{}", mapPostgresToGoType("custom_type", false))
	assert.Equal(t, "interface{}", mapPostgresToGoType("custom_type", true))
}

// Mock tests for database operations would require a test database
// For now, we'll test the business logic parts

func TestIntrospectorService_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a service with nil pool
	config := ServiceConfig{
		Pool:   nil,
		Schema: "test",
		Logger: logger,
	}

	service := NewIntrospectorService(config)

	// This should not panic even with nil pool
	// We need to handle the nil case in the Close method
	err := service.Close()
	assert.NoError(t, err)
}

// Integration tests would require a real database connection
// These tests focus on the logic that can be tested without a database

func TestIntrospectorService_DatabaseOperations_Mock(t *testing.T) {
	// This is where we would add tests with a mock database
	// or test database container for full integration testing
	t.Skip("Database integration tests require test database setup")
}

// Benchmark tests for performance analysis
func BenchmarkMapPostgresToGoType(b *testing.B) {
	types := []string{
		"integer", "bigint", "text", "boolean", "timestamp",
		"uuid", "json", "decimal", "real", "bytea",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pgType := range types {
			mapPostgresToGoType(pgType, i%2 == 0)
		}
	}
}

func TestServiceConfig_Validation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name     string
		config   ServiceConfig
		expected string
	}{
		{
			name: "valid config with schema",
			config: ServiceConfig{
				Pool:   nil,
				Schema: "custom_schema",
				Logger: logger,
			},
			expected: "custom_schema",
		},
		{
			name: "valid config without schema defaults to public",
			config: ServiceConfig{
				Pool:   nil,
				Schema: "",
				Logger: logger,
			},
			expected: "public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewIntrospectorService(tt.config)
			assert.Equal(t, tt.expected, service.schema)
			assert.NotNil(t, service.logger)
		})
	}
}

// Test that demonstrates the enhanced logging functionality
func TestIntrospectorService_Logging(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	config := ServiceConfig{
		Pool:   nil,
		Schema: "test",
		Logger: logger,
	}

	service := NewIntrospectorService(config)

	// Verify the service has the logger configured
	assert.NotNil(t, service.logger)

	// The actual logging behavior would be tested with a mock logger
	// in a more comprehensive test suite
}

// Example of how to structure integration tests
func ExampleIntrospectorService_integration() {
	// This example shows how you would set up integration tests
	// with a real database connection

	/*
		// Setup test database
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, "postgres://test:test@localhost/testdb")
		if err != nil {
			panic(err)
		}
		defer pool.Close()

		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		config := ServiceConfig{
			Pool:   pool,
			Schema: "public",
			Logger: logger,
		}

		service := NewIntrospectorService(config)

		// Test introspection
		schema, err := service.IntrospectSchema(ctx, []string{"users"})
		if err != nil {
			panic(err)
		}

		fmt.Printf("Found %d tables\n", len(schema.Tables))
	*/
}

// Test helper functions
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestCreateTestLogger(t *testing.T) {
	logger := createTestLogger()
	assert.NotNil(t, logger)
}
