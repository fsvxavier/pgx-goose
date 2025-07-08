package observability

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/interfaces"
)

// StructuredLogger implements interfaces.Logger with slog
type StructuredLogger struct {
	logger *slog.Logger
	attrs  []slog.Attr
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level slog.Level, component string) interfaces.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Add timestamp formatting
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format(time.RFC3339))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &StructuredLogger{
		logger: logger,
		attrs: []slog.Attr{
			slog.String("component", component),
			slog.String("service", "pgx-goose"),
		},
	}
}

func (l *StructuredLogger) Info(msg string, args ...interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, l.attrs...)
}

func (l *StructuredLogger) Error(msg string, args ...interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, l.attrs...)
}

func (l *StructuredLogger) Debug(msg string, args ...interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, l.attrs...)
}

func (l *StructuredLogger) Warn(msg string, args ...interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, l.attrs...)
}

func (l *StructuredLogger) With(key string, value interface{}) interfaces.Logger {
	newAttrs := make([]slog.Attr, len(l.attrs)+1)
	copy(newAttrs, l.attrs)
	newAttrs[len(l.attrs)] = slog.Any(key, value)

	return &StructuredLogger{
		logger: l.logger,
		attrs:  newAttrs,
	}
}

// MetricsCollector implements interfaces.MetricsCollector
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]interface{}
	logger  interfaces.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger interfaces.Logger) interfaces.MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]interface{}),
		logger:  logger,
	}
}

func (m *MetricsCollector) IncrementCounter(name string, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(name, labels)
	if current, exists := m.metrics[key]; exists {
		if counter, ok := current.(int64); ok {
			m.metrics[key] = counter + 1
		}
	} else {
		m.metrics[key] = int64(1)
	}

	m.logger.Debug("Counter incremented",
		"metric", name,
		"labels", labels,
		"value", m.metrics[key])
}

func (m *MetricsCollector) RecordDuration(name string, duration float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(name, labels)
	m.metrics[key] = duration

	m.logger.Debug("Duration recorded",
		"metric", name,
		"duration_ms", duration,
		"labels", labels)
}

func (m *MetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.buildKey(name, labels)
	m.metrics[key] = value

	m.logger.Debug("Gauge recorded",
		"metric", name,
		"value", value,
		"labels", labels)
}

func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range m.metrics {
		result[k] = v
	}
	return result
}

func (m *MetricsCollector) buildKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += "," + k + "=" + v
	}
	return key
}

// Observer combines logger and metrics for comprehensive observability
type Observer struct {
	Logger  interfaces.Logger
	Metrics interfaces.MetricsCollector
}

// NewObserver creates a new observer with logger and metrics
func NewObserver(component string, logLevel slog.Level) *Observer {
	logger := NewStructuredLogger(logLevel, component)
	metrics := NewMetricsCollector(logger)

	return &Observer{
		Logger:  logger,
		Metrics: metrics,
	}
}

// TimedOperation measures operation duration and logs it
func (o *Observer) TimedOperation(name string, labels map[string]string, operation func() error) error {
	start := time.Now()

	o.Logger.Info("Operation started", "operation", name, "labels", labels)

	err := operation()
	duration := float64(time.Since(start).Nanoseconds()) / 1e6 // Convert to milliseconds

	if err != nil {
		o.Logger.Error("Operation failed", "operation", name, "error", err, "duration_ms", duration)
		o.Metrics.IncrementCounter("operation_failures", map[string]string{
			"operation": name,
		})
	} else {
		o.Logger.Info("Operation completed", "operation", name, "duration_ms", duration)
		o.Metrics.IncrementCounter("operation_successes", map[string]string{
			"operation": name,
		})
	}

	o.Metrics.RecordDuration("operation_duration", duration, map[string]string{
		"operation": name,
	})

	return err
}
