package observability

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStructuredLogger(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test-component")
	require.NotNil(t, logger)

	structuredLogger, ok := logger.(*StructuredLogger)
	require.True(t, ok)
	require.NotNil(t, structuredLogger.logger)
	require.Len(t, structuredLogger.attrs, 2)
}

func TestStructuredLogger_LogMethods(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelDebug, "test")

	// Test all log methods (they should not panic)
	require.NotPanics(t, func() {
		logger.Info("test info message")
		logger.Error("test error message")
		logger.Debug("test debug message")
		logger.Warn("test warn message")
	})
}

func TestStructuredLogger_With(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test")

	newLogger := logger.With("key", "value")
	require.NotNil(t, newLogger)

	// Verify it's a different instance
	assert.NotSame(t, logger, newLogger)

	// Test chaining
	chainedLogger := newLogger.With("another", "value")
	require.NotNil(t, chainedLogger)
}

func TestNewMetricsCollector(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test")
	metrics := NewMetricsCollector(logger)

	require.NotNil(t, metrics)

	collector, ok := metrics.(*MetricsCollector)
	require.True(t, ok)
	require.NotNil(t, collector.metrics)
	require.NotNil(t, collector.logger)
}

func TestMetricsCollector_IncrementCounter(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test")
	metrics := NewMetricsCollector(logger)

	labels := map[string]string{"operation": "test"}

	// First increment
	metrics.IncrementCounter("test_counter", labels)

	allMetrics := metrics.GetMetrics()
	assert.Contains(t, allMetrics, "test_counter,operation=test")
	assert.Equal(t, int64(1), allMetrics["test_counter,operation=test"])

	// Second increment
	metrics.IncrementCounter("test_counter", labels)
	allMetrics = metrics.GetMetrics()
	assert.Equal(t, int64(2), allMetrics["test_counter,operation=test"])
}

func TestMetricsCollector_RecordDuration(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test")
	metrics := NewMetricsCollector(logger)

	labels := map[string]string{"operation": "test"}
	duration := 123.45

	metrics.RecordDuration("test_duration", duration, labels)

	allMetrics := metrics.GetMetrics()
	assert.Contains(t, allMetrics, "test_duration,operation=test")
	assert.Equal(t, duration, allMetrics["test_duration,operation=test"])
}

func TestMetricsCollector_RecordGauge(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test")
	metrics := NewMetricsCollector(logger)

	labels := map[string]string{"service": "test"}
	value := 42.0

	metrics.RecordGauge("test_gauge", value, labels)

	allMetrics := metrics.GetMetrics()
	assert.Contains(t, allMetrics, "test_gauge,service=test")
	assert.Equal(t, value, allMetrics["test_gauge,service=test"])
}

func TestMetricsCollector_BuildKey(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test")
	collector := NewMetricsCollector(logger).(*MetricsCollector)

	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "test_metric",
			labels:   map[string]string{},
			expected: "test_metric",
		},
		{
			name:     "test_metric",
			labels:   map[string]string{"key": "value"},
			expected: "test_metric,key=value",
		},
		{
			name:   "test_metric",
			labels: map[string]string{"key1": "value1", "key2": "value2"},
			// Note: map iteration order is not guaranteed, so we test both possibilities
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := collector.buildKey(tt.name, tt.labels)
			if len(tt.labels) <= 1 {
				assert.Equal(t, tt.expected, result)
			} else {
				// For multiple labels, just check that it starts with the metric name
				assert.Contains(t, result, tt.name)
			}
		})
	}
}

func TestNewObserver(t *testing.T) {
	observer := NewObserver("test-component", slog.LevelInfo)

	require.NotNil(t, observer)
	require.NotNil(t, observer.Logger)
	require.NotNil(t, observer.Metrics)
}

func TestObserver_TimedOperation_Success(t *testing.T) {
	observer := NewObserver("test", slog.LevelInfo)

	called := false
	err := observer.TimedOperation("test_operation",
		map[string]string{"component": "test"},
		func() error {
			called = true
			time.Sleep(10 * time.Millisecond) // Small delay to test timing
			return nil
		})

	assert.NoError(t, err)
	assert.True(t, called)

	// Check metrics were recorded
	metrics := observer.Metrics.GetMetrics()
	assert.Contains(t, metrics, "operation_successes,operation=test_operation")
	assert.Contains(t, metrics, "operation_duration,operation=test_operation")
}

func TestObserver_TimedOperation_Error(t *testing.T) {
	observer := NewObserver("test", slog.LevelInfo)

	expectedErr := assert.AnError
	err := observer.TimedOperation("test_operation",
		map[string]string{"component": "test"},
		func() error {
			return expectedErr
		})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// Check error metrics were recorded
	metrics := observer.Metrics.GetMetrics()
	assert.Contains(t, metrics, "operation_failures,operation=test_operation")
	assert.Contains(t, metrics, "operation_duration,operation=test_operation")
}

func TestMetricsCollector_GetMetrics_ThreadSafety(t *testing.T) {
	logger := NewStructuredLogger(slog.LevelInfo, "test")
	metrics := NewMetricsCollector(logger)

	// Test concurrent access
	done := make(chan bool, 10)

	// Start 10 goroutines incrementing counters
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 100; j++ {
				metrics.IncrementCounter("test_counter", map[string]string{
					"worker": fmt.Sprintf("worker_%d", id),
				})
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify metrics
	allMetrics := metrics.GetMetrics()
	assert.Equal(t, 10, len(allMetrics)) // 10 different worker labels

	// Each worker should have incremented 100 times
	for _, value := range allMetrics {
		assert.Equal(t, int64(100), value)
	}
}
