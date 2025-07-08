package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/container"
)

func main() {
	// Configuration for example
	cfg := &config.Config{
		DSN:       "postgres://user:password@localhost:5432/pgx_goose_example",
		Schema:    "public",
		OutputDir: "./generated",
		OutputDirs: config.OutputDirs{
			Base:       "./generated",
			Models:     "./generated/models",
			Interfaces: "./generated/interfaces",
			Repos:      "./generated/repositories",
			Tests:      "./generated/tests",
			Mocks:      "./generated/mocks",
		},
		WithTests: true,
		Parallel: config.ParallelConfig{
			Enabled: true,
			Workers: 4,
		},
		TemplateOptimization: config.TemplateOptimizationConfig{
			Enabled:   true,
			CacheSize: 100,
		},
	}

	// Create dependency container
	container, err := container.NewContainer(cfg)
	if err != nil {
		log.Printf("Failed to create container: %v", err)
		log.Println("This is expected without a real database connection")
		demonstrateContainerUsage()
		return
	}
	defer container.Close()

	// Demonstrate container usage
	demonstrateWithContainer(container)
}

func demonstrateContainerUsage() {
	fmt.Println("=== Dependency Injection Container Example ===")
	fmt.Println()

	fmt.Println("🏗️  Container Pattern Benefits:")
	fmt.Println("  ✅ Centralized dependency management")
	fmt.Println("  ✅ Proper lifecycle management")
	fmt.Println("  ✅ Easy testing with mocks")
	fmt.Println("  ✅ Clean separation of concerns")
	fmt.Println()

	fmt.Println("📋 Container Components:")
	fmt.Println("  🔧 Configuration Management")
	fmt.Println("  📊 Structured Logging")
	fmt.Println("  📈 Metrics Collection")
	fmt.Println("  🗃️  Database Pool")
	fmt.Println("  🔍 Schema Introspector")
	fmt.Println("  ⚡ Template Optimizer")
	fmt.Println("  🏭 Code Generator")
	fmt.Println()

	fmt.Println("🔄 Dependency Flow:")
	fmt.Println("  Config → Logger → Metrics → Database → Introspector → Generator")
	fmt.Println()

	fmt.Println("💡 Usage Example:")
	fmt.Println(`
  // Create container with dependencies
  container, err := container.NewContainer(config)
  if err != nil {
      return err
  }
  defer container.Close()

  // Use components
  logger := container.GetLogger()
  generator := container.GetGenerator()
  
  // Generate code
  err = generator.Generate(ctx, schema, outputPath)
	`)
}

func demonstrateWithContainer(c *container.Container) {
	ctx := context.Background()

	fmt.Println("=== Container Components Demo ===")
	fmt.Println()

	// Demonstrate logger
	logger := c.GetLogger()
	if logger != nil {
		logger.Info("Container initialized successfully")
		fmt.Println("✅ Logger: Working")
	}

	// Demonstrate metrics
	metrics := c.GetMetrics()
	if metrics != nil {
		metrics.IncrementCounter("demo_requests", map[string]string{
			"type": "example",
		})
		metrics.RecordDuration("demo_duration", 1.5, nil)
		fmt.Println("✅ Metrics: Working")
	}

	// Demonstrate health check
	err := c.Health(ctx)
	if err != nil {
		fmt.Printf("❌ Health Check: %v\n", err)
	} else {
		fmt.Println("✅ Health Check: Passed")
	}

	// Demonstrate components
	config := c.GetConfig()
	fmt.Printf("✅ Config: DSN masked, Output: %s\n", config.OutputDir)

	introspector := c.GetIntrospector()
	if introspector != nil {
		fmt.Println("✅ Introspector: Available")
	}

	generator := c.GetGenerator()
	if generator != nil {
		fmt.Println("✅ Generator: Available")
		metrics := generator.GetMetrics()
		fmt.Printf("   - Metrics: %+v\n", metrics)
	}

	optimizer := c.GetTemplateOptimizer()
	if optimizer != nil {
		fmt.Println("✅ Template Optimizer: Available")
		stats := optimizer.GetCacheStats()
		fmt.Printf("   - Cache Stats: %+v\n", stats)
	}

	fmt.Println()
	fmt.Println("🚀 Container demonstration completed!")
}

// Example of testing with dependency injection
func ExampleTestWithContainer() {
	// This example shows how the container makes testing easier

	// 1. Create test configuration
	testConfig := &config.Config{
		DSN:       "postgres://test:test@localhost:5432/test_db",
		Schema:    "test_schema",
		OutputDir: "/tmp/test_output",
		WithTests: true,
	}

	// 2. Create container (this would use mocks in real tests)
	container, err := container.NewContainer(testConfig)
	if err != nil {
		fmt.Printf("Test setup failed: %v\n", err)
		return
	}
	defer container.Close()

	// 3. Test individual components
	logger := container.GetLogger()
	if logger != nil {
		logger.Info("Test execution started")
	}

	// 4. Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = container.Health(ctx)
	fmt.Printf("Health check result: %v\n", err)

	// 5. Test metrics
	metrics := container.GetMetrics()
	if metrics != nil {
		metrics.IncrementCounter("test_counter", nil)
		allMetrics := metrics.GetMetrics()
		fmt.Printf("Test metrics: %+v\n", allMetrics)
	}

	fmt.Println("✅ Test demonstration completed")
}
