package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/interfaces"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// Mock implementations for testing
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.logs = append(m.logs, "INFO: "+msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.logs = append(m.logs, "ERROR: "+msg)
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.logs = append(m.logs, "DEBUG: "+msg)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.logs = append(m.logs, "WARN: "+msg)
}

func (m *mockLogger) With(key string, value interface{}) interfaces.Logger {
	return m
}

type mockMetrics struct {
	counters  map[string]int
	durations map[string]float64
	gauges    map[string]float64
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{
		counters:  make(map[string]int),
		durations: make(map[string]float64),
		gauges:    make(map[string]float64),
	}
}

func (m *mockMetrics) IncrementCounter(name string, labels map[string]string) {
	m.counters[name]++
}

func (m *mockMetrics) RecordDuration(name string, duration float64, labels map[string]string) {
	m.durations[name] = duration
}

func (m *mockMetrics) RecordGauge(name string, value float64, labels map[string]string) {
	m.gauges[name] = value
}

func (m *mockMetrics) GetMetrics() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m.counters {
		result[k] = v
	}
	for k, v := range m.durations {
		result[k] = v
	}
	for k, v := range m.gauges {
		result[k] = v
	}
	return result
}

type mockTemplateOptimizer struct {
	cacheStats interfaces.CacheStats
}

func (m *mockTemplateOptimizer) GetTemplate(name, content string) (interfaces.CompiledTemplate, error) {
	return nil, nil
}

func (m *mockTemplateOptimizer) ExecuteTemplate(template interfaces.CompiledTemplate, data interface{}) ([]byte, error) {
	return nil, nil
}

func (m *mockTemplateOptimizer) ClearCache() {}

func (m *mockTemplateOptimizer) PrecompileTemplates(templates map[string]string) error {
	return nil
}

func (m *mockTemplateOptimizer) GetCacheStats() interfaces.CacheStats {
	return m.cacheStats
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		OutputDir: "/tmp/test",
		Parallel: config.ParallelConfig{
			Workers: 4,
		},
	}

	gen := New(cfg)

	assert.NotNil(t, gen)
	assert.Equal(t, cfg, gen.config)
	assert.Equal(t, 4, gen.generationStats.ParallelWorkers)
}

func TestNewWithDependencies(t *testing.T) {
	cfg := &config.Config{
		OutputDir: "/tmp/test",
		Parallel: config.ParallelConfig{
			Workers: 2,
		},
	}

	logger := &mockLogger{}
	metrics := newMockMetrics()
	optimizer := &mockTemplateOptimizer{}

	gen := NewWithDependencies(cfg, logger, metrics, optimizer)

	assert.NotNil(t, gen)
	assert.IsType(t, &Generator{}, gen)

	generator := gen.(*Generator)
	assert.Equal(t, cfg, generator.config)
	assert.Equal(t, logger, generator.logger)
	assert.Equal(t, metrics, generator.metrics)
	assert.Equal(t, optimizer, generator.templateOptimizer)
}

func TestGenerator_SetTemplateOptimizer(t *testing.T) {
	gen := New(&config.Config{})
	optimizer := &mockTemplateOptimizer{}

	gen.SetTemplateOptimizer(optimizer)

	assert.Equal(t, optimizer, gen.templateOptimizer)
}

func TestGenerator_GetMetrics(t *testing.T) {
	gen := New(&config.Config{
		Parallel: config.ParallelConfig{Workers: 3},
	})

	metrics := gen.GetMetrics()

	assert.Equal(t, 3, metrics.ParallelWorkers)
	assert.Equal(t, 0, metrics.TablesProcessed)
	assert.Equal(t, 0, metrics.FilesGenerated)
	assert.Equal(t, float64(0), metrics.Duration)
}

func TestGenerator_GetMetricsWithOptimizer(t *testing.T) {
	gen := New(&config.Config{})
	optimizer := &mockTemplateOptimizer{
		cacheStats: interfaces.CacheStats{
			HitRatio: 0.85,
		},
	}
	gen.SetTemplateOptimizer(optimizer)

	metrics := gen.GetMetrics()

	assert.Equal(t, 0.85, metrics.CacheHitRatio)
}

func TestGenerator_CreateOutputDirectories(t *testing.T) {
	tempDir := t.TempDir()

	gen := New(&config.Config{
		OutputDir: tempDir,
	})

	err := gen.createOutputDirectories()
	require.NoError(t, err)

	// Check if directories were created
	expectedDirs := []string{
		tempDir,
		filepath.Join(tempDir, "models"),
		filepath.Join(tempDir, "interfaces"),
		filepath.Join(tempDir, "repositories"),
		filepath.Join(tempDir, "tests"),
		filepath.Join(tempDir, "mocks"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	}
}

func TestGenerator_GenerateSequential(t *testing.T) {
	tempDir := t.TempDir()

	logger := &mockLogger{}
	metrics := newMockMetrics()

	gen := NewWithDependencies(&config.Config{
		OutputDir: tempDir,
		WithTests: true,
		Parallel: config.ParallelConfig{
			Enabled: false,
		},
	}, logger, metrics, nil).(*Generator)

	schema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", GoType: "int", IsPrimaryKey: true},
					{Name: "name", GoType: "string"},
				},
			},
		},
	}

	err := gen.Generate(context.Background(), schema, "")
	require.NoError(t, err)

	// Check if files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "models", "users.go"),
		filepath.Join(tempDir, "interfaces", "users_repository.go"),
		filepath.Join(tempDir, "repositories", "users_repository.go"),
		filepath.Join(tempDir, "tests", "users_test.go"),
	}

	for _, file := range expectedFiles {
		_, err := os.Stat(file)
		assert.NoError(t, err, "File should exist: %s", file)
	}

	// Check metrics
	assert.Contains(t, metrics.durations, "generation_duration")
	assert.True(t, metrics.durations["generation_duration"] > 0)

	// Check logs
	assert.Contains(t, logger.logs, "INFO: Starting code generation")
	assert.Contains(t, logger.logs, "INFO: Using sequential generation")
	assert.Contains(t, logger.logs, "INFO: Code generation completed")
}

func TestGenerator_GenerateParallel(t *testing.T) {
	tempDir := t.TempDir()

	logger := &mockLogger{}
	metrics := newMockMetrics()

	gen := NewWithDependencies(&config.Config{
		OutputDir: tempDir,
		WithTests: false,
		Parallel: config.ParallelConfig{
			Enabled: true,
			Workers: 2,
		},
	}, logger, metrics, nil).(*Generator)

	schema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", GoType: "int"},
				},
			},
			{
				Name: "posts",
				Columns: []introspector.Column{
					{Name: "id", GoType: "int"},
					{Name: "title", GoType: "string"},
				},
			},
		},
	}

	err := gen.Generate(context.Background(), schema, "")
	require.NoError(t, err)

	// Check logs for parallel execution
	assert.Contains(t, logger.logs, "INFO: Using parallel generation")

	// Check if both table files were created
	userFiles := []string{
		filepath.Join(tempDir, "models", "users.go"),
		filepath.Join(tempDir, "interfaces", "users_repository.go"),
		filepath.Join(tempDir, "repositories", "users_repository.go"),
	}

	postFiles := []string{
		filepath.Join(tempDir, "models", "posts.go"),
		filepath.Join(tempDir, "interfaces", "posts_repository.go"),
		filepath.Join(tempDir, "repositories", "posts_repository.go"),
	}

	for _, file := range append(userFiles, postFiles...) {
		_, err := os.Stat(file)
		assert.NoError(t, err, "File should exist: %s", file)
	}
}

func TestGenerator_ExecuteTemplate(t *testing.T) {
	gen := New(&config.Config{})

	template := `package {{.Package}}

type {{toPascalCase .Name}} struct {
	ID int
}`

	data := map[string]interface{}{
		"Package": "models",
		"Name":    "user_profile",
	}

	result, err := gen.executeTemplate(template, data)
	require.NoError(t, err)

	expected := `package models

type UserProfile struct {
	ID int
}`

	assert.Equal(t, expected, result)
}

func TestGenerator_ExecuteTemplateError(t *testing.T) {
	gen := New(&config.Config{})

	// Invalid template syntax
	template := `package {{.Package`

	data := map[string]interface{}{
		"Package": "models",
	}

	_, err := gen.executeTemplate(template, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_profile", "UserProfile"},
		{"user_profile_setting", "UserProfileSetting"},
		{"", ""},
		{"a", "A"},
		{"a_b_c", "ABC"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := toPascalCase(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGenerator_GetGenerationMode(t *testing.T) {
	tests := []struct {
		name           string
		parallelConfig config.ParallelConfig
		expected       string
	}{
		{
			name: "parallel enabled",
			parallelConfig: config.ParallelConfig{
				Enabled: true,
			},
			expected: "parallel",
		},
		{
			name: "parallel disabled",
			parallelConfig: config.ParallelConfig{
				Enabled: false,
			},
			expected: "sequential",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gen := New(&config.Config{
				Parallel: test.parallelConfig,
			})

			result := gen.getGenerationMode()
			assert.Equal(t, test.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkGenerator_Generate(b *testing.B) {
	tempDir := b.TempDir()

	gen := New(&config.Config{
		OutputDir: tempDir,
		WithTests: false,
		Parallel: config.ParallelConfig{
			Enabled: false,
		},
	})

	schema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "benchmark_table",
				Columns: []introspector.Column{
					{Name: "id", GoType: "int"},
					{Name: "name", GoType: "string"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.Generate(context.Background(), schema, "")
	}
}

func BenchmarkToPascalCase(b *testing.B) {
	input := "user_profile_setting"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = toPascalCase(input)
	}
}

func BenchmarkGenerator_ExecuteTemplate(b *testing.B) {
	gen := New(&config.Config{})

	template := `package {{.Package}}

type {{toPascalCase .Name}} struct {
	ID int
}`

	data := map[string]interface{}{
		"Package": "models",
		"Name":    "user_profile",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.executeTemplate(template, data)
	}
}
