package interfaces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// Test type implementations
func TestGenerationMetrics(t *testing.T) {
	metrics := GenerationMetrics{
		TablesProcessed:   5,
		FilesGenerated:    20,
		ErrorsCount:       1,
		Duration:          2.5,
		ParallelWorkers:   4,
		CacheHitRatio:     0.85,
		TemplatesCompiled: 10,
	}

	assert.Equal(t, 5, metrics.TablesProcessed)
	assert.Equal(t, 20, metrics.FilesGenerated)
	assert.Equal(t, 1, metrics.ErrorsCount)
	assert.Equal(t, 2.5, metrics.Duration)
	assert.Equal(t, 4, metrics.ParallelWorkers)
	assert.Equal(t, 0.85, metrics.CacheHitRatio)
	assert.Equal(t, 10, metrics.TemplatesCompiled)
}

func TestCacheStats(t *testing.T) {
	stats := CacheStats{
		Hits:      100,
		Misses:    20,
		Evictions: 5,
		Size:      50,
		MaxSize:   100,
		HitRatio:  0.83,
	}

	assert.Equal(t, int64(100), stats.Hits)
	assert.Equal(t, int64(20), stats.Misses)
	assert.Equal(t, int64(5), stats.Evictions)
	assert.Equal(t, 50, stats.Size)
	assert.Equal(t, 100, stats.MaxSize)
	assert.Equal(t, 0.83, stats.HitRatio)
}

func TestPoolStats(t *testing.T) {
	stats := PoolStats{
		AcquireCount:         1000,
		AcquireDuration:      1.5,
		AcquiredConns:        5,
		CanceledAcquireCount: 2,
		ConstructingConns:    1,
		EmptyAcquireCount:    10,
		IdleConns:            3,
		MaxConns:             10,
		TotalConns:           8,
	}

	assert.Equal(t, int64(1000), stats.AcquireCount)
	assert.Equal(t, 1.5, stats.AcquireDuration)
	assert.Equal(t, int32(5), stats.AcquiredConns)
	assert.Equal(t, int64(2), stats.CanceledAcquireCount)
	assert.Equal(t, int32(1), stats.ConstructingConns)
	assert.Equal(t, int64(10), stats.EmptyAcquireCount)
	assert.Equal(t, int32(3), stats.IdleConns)
	assert.Equal(t, int32(10), stats.MaxConns)
	assert.Equal(t, int32(8), stats.TotalConns)
}

// Mock implementations for interface testing
type mockConfigLoader struct {
	loadError     error
	validateError error
	config        *config.Config
}

func (m *mockConfigLoader) LoadFromFile(filePath string) (*config.Config, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	return m.config, nil
}

func (m *mockConfigLoader) Validate() error {
	return m.validateError
}

func (m *mockConfigLoader) ApplyDefaults() {
	// Mock implementation
}

type mockSchemaIntrospector struct {
	schema     *introspector.Schema
	tables     []string
	closeError error
}

func (m *mockSchemaIntrospector) IntrospectSchema(ctx context.Context, tables []string) (*introspector.Schema, error) {
	return m.schema, nil
}

func (m *mockSchemaIntrospector) GetAllTables(ctx context.Context) ([]string, error) {
	return m.tables, nil
}

func (m *mockSchemaIntrospector) Close() error {
	return m.closeError
}

type mockCodeGenerator struct {
	generateError error
	metrics       GenerationMetrics
}

func (m *mockCodeGenerator) Generate(ctx context.Context, schema *introspector.Schema, outputPath string) error {
	return m.generateError
}

func (m *mockCodeGenerator) SetTemplateOptimizer(optimizer TemplateOptimizer) {
	// Mock implementation
}

func (m *mockCodeGenerator) GetMetrics() GenerationMetrics {
	return m.metrics
}

type mockTemplateOptimizer struct {
	cacheStats CacheStats
}

func (m *mockTemplateOptimizer) GetTemplate(name, content string) (CompiledTemplate, error) {
	return &mockCompiledTemplate{name: name}, nil
}

func (m *mockTemplateOptimizer) ExecuteTemplate(template CompiledTemplate, data interface{}) ([]byte, error) {
	return []byte("mock template output"), nil
}

func (m *mockTemplateOptimizer) ClearCache() {
	// Mock implementation
}

func (m *mockTemplateOptimizer) PrecompileTemplates(templates map[string]string) error {
	return nil
}

func (m *mockTemplateOptimizer) GetCacheStats() CacheStats {
	return m.cacheStats
}

type mockCompiledTemplate struct {
	name string
}

func (m *mockCompiledTemplate) Execute(data interface{}) ([]byte, error) {
	return []byte("mock execution"), nil
}

func (m *mockCompiledTemplate) Name() string {
	return m.name
}

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

func (m *mockLogger) With(key string, value interface{}) Logger {
	return m
}

type mockMetricsCollector struct {
	counters  map[string]int
	durations map[string]float64
	gauges    map[string]float64
}

func newMockMetricsCollector() *mockMetricsCollector {
	return &mockMetricsCollector{
		counters:  make(map[string]int),
		durations: make(map[string]float64),
		gauges:    make(map[string]float64),
	}
}

func (m *mockMetricsCollector) IncrementCounter(name string, labels map[string]string) {
	m.counters[name]++
}

func (m *mockMetricsCollector) RecordDuration(name string, duration float64, labels map[string]string) {
	m.durations[name] = duration
}

func (m *mockMetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
	m.gauges[name] = value
}

func (m *mockMetricsCollector) GetMetrics() map[string]interface{} {
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

type mockDatabasePool struct {
	pingError error
	stats     PoolStats
}

func (m *mockDatabasePool) Ping(ctx context.Context) error {
	return m.pingError
}

func (m *mockDatabasePool) Query(ctx context.Context, sql string, args ...interface{}) (QueryResult, error) {
	return &mockQueryResult{}, nil
}

func (m *mockDatabasePool) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	return &mockRow{}
}

func (m *mockDatabasePool) Close() {
	// Mock implementation
}

func (m *mockDatabasePool) Stats() PoolStats {
	return m.stats
}

type mockQueryResult struct {
	nextCount int
}

func (m *mockQueryResult) Next() bool {
	m.nextCount++
	return m.nextCount <= 1
}

func (m *mockQueryResult) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockQueryResult) Close() {
	// Mock implementation
}

func (m *mockQueryResult) Err() error {
	return nil
}

type mockRow struct{}

func (m *mockRow) Scan(dest ...interface{}) error {
	return nil
}

// Interface compliance tests
func TestConfigLoaderInterface(t *testing.T) {
	var loader ConfigLoader = &mockConfigLoader{
		config: &config.Config{
			DSN:       "test://localhost",
			Schema:    "test",
			OutputDir: "/tmp/test",
		},
	}

	config, err := loader.LoadFromFile("test.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	err = loader.Validate()
	assert.NoError(t, err)

	loader.ApplyDefaults()
}

func TestSchemaIntrospectorInterface(t *testing.T) {
	var introspector SchemaIntrospector = &mockSchemaIntrospector{
		schema: &introspector.Schema{
			Tables: []introspector.Table{
				{Name: "users"},
				{Name: "posts"},
			},
		},
		tables: []string{"users", "posts"},
	}

	ctx := context.Background()

	schema, err := introspector.IntrospectSchema(ctx, []string{"users"})
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Len(t, schema.Tables, 2)

	tables, err := introspector.GetAllTables(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"users", "posts"}, tables)

	err = introspector.Close()
	assert.NoError(t, err)
}

func TestCodeGeneratorInterface(t *testing.T) {
	var generator CodeGenerator = &mockCodeGenerator{
		metrics: GenerationMetrics{
			TablesProcessed: 2,
			FilesGenerated:  8,
			Duration:        1.5,
		},
	}

	ctx := context.Background()
	schema := &introspector.Schema{}

	err := generator.Generate(ctx, schema, "/tmp/output")
	assert.NoError(t, err)

	generator.SetTemplateOptimizer(&mockTemplateOptimizer{})

	metrics := generator.GetMetrics()
	assert.Equal(t, 2, metrics.TablesProcessed)
	assert.Equal(t, 8, metrics.FilesGenerated)
	assert.Equal(t, 1.5, metrics.Duration)
}

func TestTemplateOptimizerInterface(t *testing.T) {
	var optimizer TemplateOptimizer = &mockTemplateOptimizer{
		cacheStats: CacheStats{
			Hits:     100,
			Misses:   20,
			HitRatio: 0.83,
		},
	}

	template, err := optimizer.GetTemplate("test", "content")
	assert.NoError(t, err)
	assert.NotNil(t, template)

	output, err := optimizer.ExecuteTemplate(template, map[string]string{"key": "value"})
	assert.NoError(t, err)
	assert.Equal(t, []byte("mock template output"), output)

	err = optimizer.PrecompileTemplates(map[string]string{"test": "content"})
	assert.NoError(t, err)

	stats := optimizer.GetCacheStats()
	assert.Equal(t, int64(100), stats.Hits)
	assert.Equal(t, int64(20), stats.Misses)
	assert.Equal(t, 0.83, stats.HitRatio)

	optimizer.ClearCache()
}

func TestLoggerInterface(t *testing.T) {
	var logger Logger = &mockLogger{}

	logger.Info("test info message")
	logger.Error("test error message")
	logger.Debug("test debug message")
	logger.Warn("test warn message")

	child := logger.With("key", "value")
	assert.NotNil(t, child)

	mockLogger := logger.(*mockLogger)
	assert.Contains(t, mockLogger.logs, "INFO: test info message")
	assert.Contains(t, mockLogger.logs, "ERROR: test error message")
	assert.Contains(t, mockLogger.logs, "DEBUG: test debug message")
	assert.Contains(t, mockLogger.logs, "WARN: test warn message")
}

func TestMetricsCollectorInterface(t *testing.T) {
	var collector MetricsCollector = newMockMetricsCollector()

	collector.IncrementCounter("test_counter", map[string]string{"label": "value"})
	collector.IncrementCounter("test_counter", nil)

	collector.RecordDuration("test_duration", 1.5, nil)
	collector.RecordGauge("test_gauge", 42.0, nil)

	metrics := collector.GetMetrics()
	assert.Equal(t, 2, metrics["test_counter"])
	assert.Equal(t, 1.5, metrics["test_duration"])
	assert.Equal(t, 42.0, metrics["test_gauge"])
}

func TestDatabasePoolInterface(t *testing.T) {
	var pool DatabasePool = &mockDatabasePool{
		stats: PoolStats{
			MaxConns:   10,
			TotalConns: 5,
		},
	}

	ctx := context.Background()

	err := pool.Ping(ctx)
	assert.NoError(t, err)

	result, err := pool.Query(ctx, "SELECT 1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	row := pool.QueryRow(ctx, "SELECT 1", nil)
	assert.NotNil(t, row)

	stats := pool.Stats()
	assert.Equal(t, int32(10), stats.MaxConns)
	assert.Equal(t, int32(5), stats.TotalConns)

	pool.Close()
}

func TestQueryResultInterface(t *testing.T) {
	var result QueryResult = &mockQueryResult{}

	assert.True(t, result.Next())
	assert.False(t, result.Next())

	err := result.Scan()
	assert.NoError(t, err)

	err = result.Err()
	assert.NoError(t, err)

	result.Close()
}

func TestRowInterface(t *testing.T) {
	var row Row = &mockRow{}

	err := row.Scan()
	assert.NoError(t, err)
}

func TestCompiledTemplateInterface(t *testing.T) {
	var template CompiledTemplate = &mockCompiledTemplate{name: "test"}

	output, err := template.Execute(map[string]string{"key": "value"})
	assert.NoError(t, err)
	assert.Equal(t, []byte("mock execution"), output)

	name := template.Name()
	assert.Equal(t, "test", name)
}

// Performance tests
func BenchmarkMetricsCollector_IncrementCounter(b *testing.B) {
	collector := newMockMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.IncrementCounter("benchmark_counter", nil)
	}
}

func BenchmarkMetricsCollector_RecordDuration(b *testing.B) {
	collector := newMockMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordDuration("benchmark_duration", float64(i), nil)
	}
}

func BenchmarkLogger_Info(b *testing.B) {
	logger := &mockLogger{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}
