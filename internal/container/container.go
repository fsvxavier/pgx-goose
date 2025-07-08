package container

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/database"
	"github.com/fsvxavier/pgx-goose/internal/generator"
	"github.com/fsvxavier/pgx-goose/internal/interfaces"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
	"github.com/fsvxavier/pgx-goose/internal/observability"
	"github.com/fsvxavier/pgx-goose/internal/performance"
)

// Container holds all application dependencies
type Container struct {
	config            *config.Config
	logger            interfaces.Logger
	metrics           interfaces.MetricsCollector
	dbPool            interfaces.DatabasePool
	introspector      interfaces.SchemaIntrospector
	templateOptimizer interfaces.TemplateOptimizer
	generator         interfaces.CodeGenerator
}

// NewContainer creates a new dependency container
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{
		config: cfg,
	}

	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return container, nil
}

// initializeServices initializes all services with proper dependency injection
func (c *Container) initializeServices() error {
	var err error

	// Initialize logger first (needed by other services)
	c.logger = observability.NewStructuredLogger(slog.LevelInfo, "pgx-goose")
	c.logger.Info("Initializing container services")

	// Initialize metrics collector (simple implementation)
	c.metrics = &enhancedMetricsCollector{
		metrics:   make(map[string]interface{}),
		startTime: time.Now(),
	}
	c.logger.Info("Metrics collector initialized")

	// Initialize database pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c.dbPool, err = database.NewPgxPoolAdapter(ctx, c.config.DSN)
	if err != nil {
		c.logger.Error("Failed to initialize database pool", "error", err)
		return fmt.Errorf("failed to initialize database pool: %w", err)
	}

	// Test database connection with retry logic
	if err := c.retryDatabaseConnection(ctx); err != nil {
		c.logger.Error("Database connection failed after retries", "error", err)
		return fmt.Errorf("failed to establish database connection: %w", err)
	}
	c.logger.Info("Database connection established")

	// Initialize introspector with dependencies
	c.introspector = &introspectorAdapter{
		introspector: introspector.New(c.config.DSN, c.config.Schema),
	}
	c.logger.Info("Schema introspector initialized")

	// Initialize template optimizer with configuration
	cacheSize := c.config.TemplateOptimization.CacheSize
	if cacheSize <= 0 {
		cacheSize = 50 // default cache size
	}
	c.templateOptimizer = performance.NewTemplateOptimizer(cacheSize, nil)
	c.logger.Info("Template optimizer initialized", "cacheSize", cacheSize)

	// Initialize generator with full dependency injection
	c.generator = generator.NewWithDependencies(
		c.config,
		c.logger,
		c.metrics,
		c.templateOptimizer,
	)
	c.logger.Info("Code generator initialized")

	c.logger.Info("All container services initialized successfully")
	return nil
}

// generatorAdapter adapts *generator.Generator to interfaces.CodeGenerator
type generatorAdapter struct {
	generator *generator.Generator
}

func (g *generatorAdapter) Generate(ctx context.Context, schema *introspector.Schema, outputPath string) error {
	return g.generator.Generate(ctx, schema, outputPath)
}

func (g *generatorAdapter) SetTemplateOptimizer(optimizer interfaces.TemplateOptimizer) {
	// Current generator doesn't support template optimizer
}

func (g *generatorAdapter) GetMetrics() interfaces.GenerationMetrics {
	return interfaces.GenerationMetrics{
		TablesProcessed: 0, // Would need to be tracked
		FilesGenerated:  0,
		ErrorsCount:     0,
		Duration:        0,
	}
}

// introspectorAdapter adapts *introspector.Introspector to interfaces.SchemaIntrospector
type introspectorAdapter struct {
	introspector *introspector.Introspector
}

func (i *introspectorAdapter) IntrospectSchema(ctx context.Context, tables []string) (*introspector.Schema, error) {
	return i.introspector.IntrospectSchema(tables)
}

func (i *introspectorAdapter) GetAllTables(ctx context.Context) ([]string, error) {
	return i.introspector.GetAllTables()
}

func (i *introspectorAdapter) Close() error {
	i.introspector.Close()
	return nil
}

// enhancedMetricsCollector provides enhanced metrics implementation
type enhancedMetricsCollector struct {
	mu        sync.RWMutex
	metrics   map[string]interface{}
	startTime time.Time
}

func (e *enhancedMetricsCollector) IncrementCounter(name string, labels map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.metrics == nil {
		e.metrics = make(map[string]interface{})
	}

	key := name
	if len(labels) > 0 {
		// Simple label handling
		for k, v := range labels {
			key += fmt.Sprintf("_%s_%s", k, v)
		}
	}

	if current, exists := e.metrics[key]; exists {
		if count, ok := current.(int); ok {
			e.metrics[key] = count + 1
		}
	} else {
		e.metrics[key] = 1
	}
}

func (e *enhancedMetricsCollector) RecordDuration(name string, duration float64, labels map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.metrics == nil {
		e.metrics = make(map[string]interface{})
	}

	key := name
	if len(labels) > 0 {
		for k, v := range labels {
			key += fmt.Sprintf("_%s_%s", k, v)
		}
	}

	e.metrics[key] = duration
}

func (e *enhancedMetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.metrics == nil {
		e.metrics = make(map[string]interface{})
	}

	key := name
	if len(labels) > 0 {
		for k, v := range labels {
			key += fmt.Sprintf("_%s_%s", k, v)
		}
	}

	e.metrics[key] = value
}

func (e *enhancedMetricsCollector) GetMetrics() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range e.metrics {
		result[k] = v
	}

	// Add uptime
	result["uptime_seconds"] = time.Since(e.startTime).Seconds()

	return result
}

// retryDatabaseConnection attempts to connect to the database with retries
func (c *Container) retryDatabaseConnection(ctx context.Context) error {
	maxRetries := 3
	retryDelay := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		if err := c.dbPool.Ping(ctx); err == nil {
			return nil
		}

		if i < maxRetries-1 {
			c.logger.Warn("Database connection failed, retrying...",
				"attempt", i+1,
				"maxRetries", maxRetries)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
}

// simpleMetricsCollector provides a basic metrics implementation
type simpleMetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]interface{}
}

func (s *simpleMetricsCollector) IncrementCounter(name string, labels map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metrics == nil {
		s.metrics = make(map[string]interface{})
	}
	if current, exists := s.metrics[name]; exists {
		if count, ok := current.(int); ok {
			s.metrics[name] = count + 1
		}
	} else {
		s.metrics[name] = 1
	}
}

func (s *simpleMetricsCollector) RecordDuration(name string, duration float64, labels map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metrics == nil {
		s.metrics = make(map[string]interface{})
	}
	s.metrics[name] = duration
}

func (s *simpleMetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.metrics == nil {
		s.metrics = make(map[string]interface{})
	}
	s.metrics[name] = value
}

func (s *simpleMetricsCollector) GetMetrics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range s.metrics {
		result[k] = v
	}
	return result
}

// GetConfig returns the configuration
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// GetLogger returns the logger
func (c *Container) GetLogger() interfaces.Logger {
	return c.logger
}

// GetMetrics returns the metrics collector
func (c *Container) GetMetrics() interfaces.MetricsCollector {
	return c.metrics
}

// GetDatabasePool returns the database pool
func (c *Container) GetDatabasePool() interfaces.DatabasePool {
	return c.dbPool
}

// GetIntrospector returns the schema introspector
func (c *Container) GetIntrospector() interfaces.SchemaIntrospector {
	return c.introspector
}

// GetTemplateOptimizer returns the template optimizer
func (c *Container) GetTemplateOptimizer() interfaces.TemplateOptimizer {
	return c.templateOptimizer
}

// GetGenerator returns the code generator
func (c *Container) GetGenerator() interfaces.CodeGenerator {
	return c.generator
}

// Close closes all resources
func (c *Container) Close() error {
	var errs []error

	if c.introspector != nil {
		if err := c.introspector.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close introspector: %w", err))
		}
	}

	if c.dbPool != nil {
		c.dbPool.Close()
	}

	if c.templateOptimizer != nil {
		c.templateOptimizer.ClearCache()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errs)
	}

	return nil
}

// Health checks the health of all services
func (c *Container) Health(ctx context.Context) error {
	// Check database connection
	if c.dbPool == nil {
		return fmt.Errorf("database health check failed: database pool is nil")
	}

	if err := c.dbPool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Log health check if logger is available
	if c.logger != nil {
		c.logger.Info("Health check passed", "service", "container")
	}

	return nil
}
