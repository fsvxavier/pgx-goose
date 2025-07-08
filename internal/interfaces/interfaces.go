package interfaces

import (
	"context"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

//go:generate mockgen -source=interfaces.go -destination=../mocks/interfaces_mock.go -package=mocks

// ConfigLoader abstracts configuration loading
type ConfigLoader interface {
	LoadFromFile(filePath string) (*config.Config, error)
	Validate() error
	ApplyDefaults()
}

// SchemaIntrospector abstracts database schema introspection
type SchemaIntrospector interface {
	IntrospectSchema(ctx context.Context, tables []string) (*introspector.Schema, error)
	GetAllTables(ctx context.Context) ([]string, error)
	Close() error
}

// CodeGenerator abstracts code generation
type CodeGenerator interface {
	Generate(ctx context.Context, schema *introspector.Schema, outputPath string) error
	SetTemplateOptimizer(optimizer TemplateOptimizer)
	GetMetrics() GenerationMetrics
}

// TemplateOptimizer abstracts template compilation and caching
type TemplateOptimizer interface {
	GetTemplate(name, content string) (CompiledTemplate, error)
	ExecuteTemplate(template CompiledTemplate, data interface{}) ([]byte, error)
	ClearCache()
	PrecompileTemplates(templates map[string]string) error
	GetCacheStats() CacheStats
}

// CompiledTemplate represents a compiled template
type CompiledTemplate interface {
	Execute(data interface{}) ([]byte, error)
	Name() string
}

// Logger abstracts structured logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	With(key string, value interface{}) Logger
}

// MetricsCollector abstracts metrics collection
type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)
	RecordDuration(name string, duration float64, labels map[string]string)
	RecordGauge(name string, value float64, labels map[string]string)
	GetMetrics() map[string]interface{}
}

// GenerationMetrics contains generation statistics
type GenerationMetrics struct {
	TablesProcessed   int
	FilesGenerated    int
	ErrorsCount       int
	Duration          float64
	ParallelWorkers   int
	CacheHitRatio     float64
	TemplatesCompiled int
}

// CacheStats contains template cache statistics
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
	MaxSize   int
	HitRatio  float64
}

// DatabasePool abstracts database connection pooling
type DatabasePool interface {
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...interface{}) (QueryResult, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) Row
	Close()
	Stats() PoolStats
}

// QueryResult abstracts database query results
type QueryResult interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close()
	Err() error
}

// Row abstracts single row results
type Row interface {
	Scan(dest ...interface{}) error
}

// PoolStats contains connection pool statistics
type PoolStats struct {
	AcquireCount         int64
	AcquireDuration      float64
	AcquiredConns        int32
	CanceledAcquireCount int64
	ConstructingConns    int32
	EmptyAcquireCount    int64
	IdleConns            int32
	MaxConns             int32
	TotalConns           int32
}
