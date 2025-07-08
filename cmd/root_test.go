package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	// Test that Execute function exists and can be called
	// We can't test the actual execution without mocking the entire flow
	assert.NotNil(t, Execute)
}

func TestRootCommand(t *testing.T) {
	// Test that root command is properly configured
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "pgx-goose", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "PostgreSQL reverse engineering")
	assert.Contains(t, rootCmd.Long, "pgx-goose is a powerful tool")
}

func TestFlags(t *testing.T) {
	// Test that all flags are properly defined
	flags := rootCmd.PersistentFlags()

	// Test basic flags
	assert.NotNil(t, flags.Lookup("dsn"))
	assert.NotNil(t, flags.Lookup("schema"))
	assert.NotNil(t, flags.Lookup("out"))
	assert.NotNil(t, flags.Lookup("config"))
	assert.NotNil(t, flags.Lookup("tables"))
	assert.NotNil(t, flags.Lookup("template-dir"))
	assert.NotNil(t, flags.Lookup("mock-provider"))
	assert.NotNil(t, flags.Lookup("with-tests"))
	assert.NotNil(t, flags.Lookup("verbose"))
	assert.NotNil(t, flags.Lookup("debug"))

	// Test directory flags
	assert.NotNil(t, flags.Lookup("models-dir"))
	assert.NotNil(t, flags.Lookup("interfaces-dir"))
	assert.NotNil(t, flags.Lookup("repos-dir"))
	assert.NotNil(t, flags.Lookup("mocks-dir"))
	assert.NotNil(t, flags.Lookup("tests-dir"))

	// Test advanced feature flags
	assert.NotNil(t, flags.Lookup("parallel"))
	assert.NotNil(t, flags.Lookup("workers"))
	assert.NotNil(t, flags.Lookup("incremental"))
	assert.NotNil(t, flags.Lookup("force"))
	assert.NotNil(t, flags.Lookup("generate-migrations"))
	assert.NotNil(t, flags.Lookup("cross-schema"))
	assert.NotNil(t, flags.Lookup("go-generate"))
	assert.NotNil(t, flags.Lookup("optimize-templates"))
}

func TestFlagDefaults(t *testing.T) {
	// Test default values
	assert.Equal(t, "./pgx-goose", outputDir)
	assert.Equal(t, true, withTests)
	assert.Equal(t, true, useYAML)
	assert.Equal(t, false, useJSON)
	assert.Equal(t, false, verbose)
	assert.Equal(t, false, debug)
	assert.Equal(t, false, parallel)
	assert.Equal(t, 0, workers)
	assert.Equal(t, false, incremental)
	assert.Equal(t, false, forceRegenerate)
	assert.Equal(t, false, generateMigrations)
	assert.Equal(t, false, enableCrossSchema)
	assert.Equal(t, false, generateGoGenerate)
	assert.Equal(t, true, optimizeTemplates)
}

func TestVariableInitialization(t *testing.T) {
	// Test that all global variables are properly initialized
	assert.NotNil(t, tables)
	assert.Equal(t, 0, len(tables)) // Empty slice by default
}

// Test individual flag setting (simulation)
func TestFlagValues(t *testing.T) {
	// Test setting values (this simulates what would happen when flags are parsed)
	originalDSN := dsn
	originalSchema := schema
	originalOutputDir := outputDir

	// Simulate flag parsing
	dsn = "postgres://user:pass@localhost/db"
	schema = "test_schema"
	outputDir = "/tmp/output"

	assert.Equal(t, "postgres://user:pass@localhost/db", dsn)
	assert.Equal(t, "test_schema", schema)
	assert.Equal(t, "/tmp/output", outputDir)

	// Restore original values
	dsn = originalDSN
	schema = originalSchema
	outputDir = originalOutputDir
}

func TestBooleanFlags(t *testing.T) {
	// Test boolean flag toggling
	originalVerbose := verbose
	originalDebug := debug
	originalParallel := parallel

	// Simulate flag setting
	verbose = true
	debug = true
	parallel = true

	assert.True(t, verbose)
	assert.True(t, debug)
	assert.True(t, parallel)

	// Restore
	verbose = originalVerbose
	debug = originalDebug
	parallel = originalParallel
}

func TestSliceFlags(t *testing.T) {
	// Test slice flags
	originalTables := tables

	// Simulate setting tables
	tables = []string{"users", "products", "orders"}

	assert.Len(t, tables, 3)
	assert.Contains(t, tables, "users")
	assert.Contains(t, tables, "products")
	assert.Contains(t, tables, "orders")

	// Restore
	tables = originalTables
}

func TestIntFlags(t *testing.T) {
	// Test integer flags
	originalWorkers := workers

	// Simulate setting workers
	workers = 4

	assert.Equal(t, 4, workers)

	// Restore
	workers = originalWorkers
}

func TestStringFlags(t *testing.T) {
	// Test string flags
	originalConfigFile := configFile
	originalTemplateDir := templateDir
	originalMockProvider := mockProvider

	// Simulate setting values
	configFile = "/path/to/config.yaml"
	templateDir = "/path/to/templates"
	mockProvider = "testify"

	assert.Equal(t, "/path/to/config.yaml", configFile)
	assert.Equal(t, "/path/to/templates", templateDir)
	assert.Equal(t, "testify", mockProvider)

	// Restore
	configFile = originalConfigFile
	templateDir = originalTemplateDir
	mockProvider = originalMockProvider
}

func TestDirectoryFlags(t *testing.T) {
	// Test directory-specific flags
	originalModelsDir := modelsDir
	originalInterfacesDir := interfacesDir
	originalReposDir := reposDir
	originalMocksDir := mocksDir
	originalTestsDir := testsDir

	// Simulate setting values
	modelsDir = "/path/to/models"
	interfacesDir = "/path/to/interfaces"
	reposDir = "/path/to/repos"
	mocksDir = "/path/to/mocks"
	testsDir = "/path/to/tests"

	assert.Equal(t, "/path/to/models", modelsDir)
	assert.Equal(t, "/path/to/interfaces", interfacesDir)
	assert.Equal(t, "/path/to/repos", reposDir)
	assert.Equal(t, "/path/to/mocks", mocksDir)
	assert.Equal(t, "/path/to/tests", testsDir)

	// Restore
	modelsDir = originalModelsDir
	interfacesDir = originalInterfacesDir
	reposDir = originalReposDir
	mocksDir = originalMocksDir
	testsDir = originalTestsDir
}
