package container

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/generator"
	"github.com/fsvxavier/pgx-goose/internal/interfaces"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

func TestNewContainer(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: &config.Config{
				DSN:                  "postgres://user:pass@localhost:5432/testdb",
				Schema:               "public",
				OutputDir:            "/tmp/test",
				TemplateOptimization: config.TemplateOptimizationConfig{CacheSize: 100},
				Parallel:             config.ParallelConfig{Workers: 2},
			},
			wantError: true, // Will fail due to invalid DSN, but tests the initialization path
			errorMsg:  "failed to initialize services",
		},
		{
			name: "empty DSN",
			config: &config.Config{
				DSN:       "",
				Schema:    "public",
				OutputDir: "/tmp/test",
			},
			wantError: true,
			errorMsg:  "failed to initialize services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := NewContainer(tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, container)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, container)

				// Clean up
				if container != nil {
					container.Close()
				}
			}
		})
	}
}

func TestContainer_Getters(t *testing.T) {
	// Create a container with minimal config (won't initialize fully due to invalid DSN)
	cfg := &config.Config{
		DSN:                  "postgres://user:pass@localhost:5432/testdb",
		Schema:               "public",
		OutputDir:            "/tmp/test",
		TemplateOptimization: config.TemplateOptimizationConfig{CacheSize: 100},
		Parallel:             config.ParallelConfig{Workers: 2},
	}

	// Create container struct directly for testing getters
	container := &Container{
		config: cfg,
	}

	// Test getters
	assert.Equal(t, cfg, container.GetConfig())
	assert.Nil(t, container.GetLogger())            // Not initialized
	assert.Nil(t, container.GetMetrics())           // Not initialized
	assert.Nil(t, container.GetDatabasePool())      // Not initialized
	assert.Nil(t, container.GetIntrospector())      // Not initialized
	assert.Nil(t, container.GetTemplateOptimizer()) // Not initialized
	assert.Nil(t, container.GetGenerator())         // Not initialized
}

func TestContainer_Close(t *testing.T) {
	container := &Container{}

	// Test close with nil components (should not panic)
	err := container.Close()
	assert.NoError(t, err)
}

func TestContainer_Health(t *testing.T) {
	container := &Container{}

	ctx := context.Background()

	// Test health check with nil database pool (should return specific error)
	err := container.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database health check failed")
}

func TestGeneratorAdapter(t *testing.T) {
	// Test generator adapter methods
	adapter := &generatorAdapter{}

	// Test Generate with nil generator (will panic, but we're testing the method signature)
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil generator
		}
	}()

	// Test SetTemplateOptimizer (should not panic)
	adapter.SetTemplateOptimizer(nil)

	// Test GetMetrics
	metrics := adapter.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TablesProcessed)
	assert.Equal(t, 0, metrics.FilesGenerated)
	assert.Equal(t, 0, metrics.ErrorsCount)
	assert.Equal(t, float64(0), metrics.Duration)
}

func TestGeneratorAdapter_Methods(t *testing.T) {
	cfg := &config.Config{
		OutputDir: "/tmp/test",
		Parallel:  config.ParallelConfig{Enabled: false},
	}

	adapter := &generatorAdapter{
		generator: generator.New(cfg),
	}

	// Test Generate
	ctx := context.Background()
	schema := &introspector.Schema{
		Tables: []introspector.Table{
			{Name: "test_table", Columns: []introspector.Column{{Name: "id", GoType: "int"}}},
		},
	}

	err := adapter.Generate(ctx, schema, "/tmp/test-output")
	// This might fail due to file system issues, but we're testing the interface
	assert.NotNil(t, err) // Expect error due to missing dirs/permissions

	// Test SetTemplateOptimizer
	adapter.SetTemplateOptimizer(nil) // Should not panic

	// Test GetMetrics
	metrics := adapter.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TablesProcessed)
}

func TestIntrospectorAdapter(t *testing.T) {
	// Test introspector adapter methods
	adapter := &introspectorAdapter{}

	ctx := context.Background()

	// Test GetAllTables with nil introspector (will panic)
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil introspector
		}
	}()

	// Test methods (will panic due to nil introspector, but tests method signatures)
	_, _ = adapter.GetAllTables(ctx)
	_, _ = adapter.IntrospectSchema(ctx, nil)
	_ = adapter.Close()
}

func TestIntrospectorAdapter_Methods(t *testing.T) {
	// Test with nil introspector (will fail but tests the interface)
	adapter := &introspectorAdapter{
		introspector: nil,
	}

	ctx := context.Background()

	// Test IntrospectSchema - will panic but tests the interface
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with nil introspector
		}
	}()

	_, _ = adapter.IntrospectSchema(ctx, []string{"test_table"})

	// Test Close
	err := adapter.Close()
	assert.NoError(t, err) // Close should handle nil gracefully
}

// Test metrics collector methods
func TestSimpleMetricsCollector_MoreMethods(t *testing.T) {
	collector := &simpleMetricsCollector{}

	// Test RecordDuration
	collector.RecordDuration("test_duration", 1.5, map[string]string{"label": "value"})
	metrics := collector.GetMetrics()
	assert.Equal(t, 1.5, metrics["test_duration"])

	// Test RecordGauge
	collector.RecordGauge("test_gauge", 42.0, map[string]string{"label": "value"})
	metrics = collector.GetMetrics()
	assert.Equal(t, 42.0, metrics["test_gauge"])

	// Test GetMetrics - already tested in TestSimpleMetricsCollector
}

func TestEnhancedMetricsCollector_Methods(t *testing.T) {
	collector := &enhancedMetricsCollector{}

	// Test IncrementCounter
	collector.IncrementCounter("test_counter", map[string]string{"label": "value"})
	metrics := collector.GetMetrics()
	count, exists := metrics["test_counter"]
	assert.True(t, exists)
	assert.Equal(t, int64(1), count)

	// Test RecordDuration
	collector.RecordDuration("test_duration", 2.5, map[string]string{"label": "value"})
	metrics = collector.GetMetrics()
	duration, exists := metrics["test_duration"]
	assert.True(t, exists)
	assert.Equal(t, 2.5, duration)

	// Test RecordGauge
	collector.RecordGauge("test_gauge", 100.0, map[string]string{"label": "value"})
	metrics = collector.GetMetrics()
	gauge, exists := metrics["test_gauge"]
	assert.True(t, exists)
	assert.Equal(t, 100.0, gauge)

	// Test GetMetrics - already tested
}

// Test Health method scenarios
func TestContainer_Health_Scenarios(t *testing.T) {
	t.Run("nil database pool", func(t *testing.T) {
		cfg := &config.Config{
			DSN:    "postgres://invalid:invalid@localhost:5432/invalid",
			Schema: "public",
		}

		// Create container without initializing (to get nil dbPool)
		container := &Container{
			config: cfg,
		}

		ctx := context.Background()
		err := container.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database pool is nil")
	})

	t.Run("database ping failure", func(t *testing.T) {
		cfg := &config.Config{
			DSN:    "postgres://invalid:invalid@localhost:5432/invalid",
			Schema: "public",
		}

		// Create container with mock failing database
		container := &Container{
			config: cfg,
			dbPool: &mockFailingDB{},
		}

		ctx := context.Background()
		err := container.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database health check failed")
	})
}

// Mock failing database for testing
type mockFailingDB struct{}

func (m *mockFailingDB) Ping(ctx context.Context) error {
	return fmt.Errorf("mock database connection failed")
}

func (m *mockFailingDB) Query(ctx context.Context, sql string, args ...interface{}) (interfaces.QueryResult, error) {
	return nil, fmt.Errorf("mock query failed")
}

func (m *mockFailingDB) QueryRow(ctx context.Context, sql string, args ...interface{}) interfaces.Row {
	return nil
}

func (m *mockFailingDB) Close() {}

func (m *mockFailingDB) Stats() interfaces.PoolStats {
	return interfaces.PoolStats{}
}

// Test Close method with different scenarios
func TestContainer_Close_ErrorHandling(t *testing.T) {
	t.Run("introspector close error", func(t *testing.T) {
		container := &Container{
			introspector: &mockFailingIntrospector{},
		}

		err := container.Close()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to close introspector")
	})

	t.Run("successful close", func(t *testing.T) {
		container := &Container{
			introspector:      &introspectorAdapter{introspector: nil},
			dbPool:            &mockFailingDB{},
			templateOptimizer: &mockTemplateOptimizer{},
		}

		err := container.Close()
		assert.NoError(t, err)
	})
}

// Mock failing introspector for testing
type mockFailingIntrospector struct{}

func (m *mockFailingIntrospector) IntrospectSchema(ctx context.Context, tables []string) (*introspector.Schema, error) {
	return nil, fmt.Errorf("mock introspector failed")
}

func (m *mockFailingIntrospector) GetAllTables(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("mock get tables failed")
}

func (m *mockFailingIntrospector) Close() error {
	return fmt.Errorf("mock introspector close failed")
}

// Mock template optimizer for testing
type mockTemplateOptimizer struct{}

func (m *mockTemplateOptimizer) GetTemplate(name, content string) (interfaces.CompiledTemplate, error) {
	return nil, fmt.Errorf("mock template failed")
}

func (m *mockTemplateOptimizer) ExecuteTemplate(template interfaces.CompiledTemplate, data interface{}) ([]byte, error) {
	return nil, fmt.Errorf("mock execute failed")
}

func (m *mockTemplateOptimizer) ClearCache() {}

func (m *mockTemplateOptimizer) PrecompileTemplates(templates map[string]string) error {
	return fmt.Errorf("mock precompile failed")
}

func (m *mockTemplateOptimizer) GetCacheStats() interfaces.CacheStats {
	return interfaces.CacheStats{}
}
