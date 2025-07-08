package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/generator"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

var (
	dsn       string
	schema    string
	outputDir string
	// New individual output directory flags
	modelsDir     string
	interfacesDir string
	reposDir      string
	mocksDir      string
	testsDir      string
	tables        []string
	configFile    string
	templateDir   string
	mockProvider  string
	withTests     bool
	useJSON       bool
	useYAML       bool
	verbose       bool
	debug         bool

	// New flags for advanced features
	parallel           bool
	workers            int
	incremental        bool
	forceRegenerate    bool
	generateMigrations bool
	enableCrossSchema  bool
	generateGoGenerate bool
	optimizeTemplates  bool
)

var rootCmd = &cobra.Command{
	Use:   "pgx-goose",
	Short: "PostgreSQL reverse engineering tool for Go code generation",
	Long: `pgx-goose is a powerful tool that performs reverse engineering on PostgreSQL databases
to automatically generate Go source code including structs, repository interfaces,
implementations, mocks, and unit tests.`,
	RunE: runGenerate,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dsn, "dsn", "", "PostgreSQL connection string")
	rootCmd.PersistentFlags().StringVar(&schema, "schema", "", "Database schema to introspect (default: public)")
	rootCmd.PersistentFlags().StringVar(&outputDir, "out", "./pgx-goose", "Output directory for generated files")

	// Individual output directory flags
	rootCmd.PersistentFlags().StringVar(&modelsDir, "models-dir", "", "Output directory for models (overrides config)")
	rootCmd.PersistentFlags().StringVar(&interfacesDir, "interfaces-dir", "", "Output directory for repository interfaces (overrides config)")
	rootCmd.PersistentFlags().StringVar(&reposDir, "repos-dir", "", "Output directory for repository implementations (overrides config)")
	rootCmd.PersistentFlags().StringVar(&mocksDir, "mocks-dir", "", "Output directory for mocks (overrides config)")
	rootCmd.PersistentFlags().StringVar(&testsDir, "tests-dir", "", "Output directory for tests (overrides config)")

	rootCmd.PersistentFlags().StringSliceVar(&tables, "tables", []string{}, "Comma-separated list of tables to process (optional)")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to configuration file (pgx-goose-conf.yaml or pgx-goose-conf.json)")
	rootCmd.PersistentFlags().StringVar(&templateDir, "template-dir", "", "Directory containing custom templates")
	rootCmd.PersistentFlags().StringVar(&mockProvider, "mock-provider", "", "Mock provider: 'testify' or 'mock'")
	rootCmd.PersistentFlags().BoolVar(&withTests, "with-tests", true, "Generate unit tests")
	rootCmd.PersistentFlags().BoolVar(&useJSON, "json", false, "Use JSON configuration format")
	rootCmd.PersistentFlags().BoolVar(&useYAML, "yaml", true, "Use YAML configuration format")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")

	// New flags for advanced features
	rootCmd.PersistentFlags().BoolVar(&parallel, "parallel", false, "Enable parallel code generation")
	rootCmd.PersistentFlags().IntVar(&workers, "workers", 0, "Number of parallel workers (0 = auto-detect)")
	rootCmd.PersistentFlags().BoolVar(&incremental, "incremental", false, "Enable incremental generation")
	rootCmd.PersistentFlags().BoolVar(&forceRegenerate, "force", false, "Force full regeneration (ignore cache)")
	rootCmd.PersistentFlags().BoolVar(&generateMigrations, "generate-migrations", false, "Generate database migrations")
	rootCmd.PersistentFlags().BoolVar(&enableCrossSchema, "cross-schema", false, "Enable cross-schema relationship detection")
	rootCmd.PersistentFlags().BoolVar(&generateGoGenerate, "go-generate", false, "Generate go:generate integration files")
	rootCmd.PersistentFlags().BoolVar(&optimizeTemplates, "optimize-templates", true, "Enable template optimization and caching")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	setupLogging()

	slog.Info("Starting pgx-goose code generation")

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	slog.Debug("Configuration loaded", "config", cfg)

	// Log specific schema information early to verify it's being read correctly
	slog.Info("Using database schema", "schema", cfg.Schema)

	// Handle go:generate integration
	if generateGoGenerate {
		return handleGoGenerateIntegration(cfg)
	}

	// Handle migration generation
	if generateMigrations {
		return handleMigrationGeneration(cfg)
	}

	// Handle cross-schema generation
	if enableCrossSchema {
		return handleCrossSchemaGeneration(cfg)
	}

	// Handle regular generation (with potential optimizations)
	return handleRegularGeneration(cfg)
}

// handleGoGenerateIntegration handles go:generate integration setup
func handleGoGenerateIntegration(cfg *config.Config) error {
	slog.Info("Setting up go:generate integration")

	// For now, use standard generation
	return handleRegularGeneration(cfg)
}

// handleMigrationGeneration handles database migration generation
func handleMigrationGeneration(cfg *config.Config) error {
	slog.Info("Generating database migrations")

	// This would need old and new schema - for now, return not implemented
	return fmt.Errorf("migration generation requires comparison between two schemas - not fully implemented yet")
}

// handleCrossSchemaGeneration handles cross-schema code generation
func handleCrossSchemaGeneration(cfg *config.Config) error {
	slog.Info("Cross-schema generation not fully implemented yet")
	return fmt.Errorf("cross-schema generation requires multi-schema configuration - not fully implemented yet")
}

// handleRegularGeneration handles regular code generation with optimizations
func handleRegularGeneration(cfg *config.Config) error {
	// Create introspector
	inspector := introspector.New(cfg.DSN, cfg.Schema)

	// Connect to database and introspect schema
	slog.Info("Connecting to database...")

	var tablesToProcess []string

	// If specific tables are requested, use them (filtered by ignore_tables)
	if len(cfg.Tables) > 0 {
		tablesToProcess = cfg.FilterTables(cfg.Tables)
		slog.Info("Processing specified tables", "tables", tablesToProcess)
	} else {
		// Let introspector get all tables, then we'll filter them afterwards
		tablesToProcess = []string{} // Empty means "get all tables"
	}

	if len(cfg.IgnoreTables) > 0 {
		slog.Info("Ignoring tables", "count", len(cfg.IgnoreTables), "tables", cfg.IgnoreTables)
	}

	schema, err := inspector.IntrospectSchema(tablesToProcess)
	if err != nil {
		return fmt.Errorf("failed to introspect database schema: %w", err)
	}

	// If we got all tables (cfg.Tables was empty), filter out ignored tables from the result
	if len(cfg.Tables) == 0 && len(cfg.IgnoreTables) > 0 {
		filteredTables := make([]introspector.Table, 0, len(schema.Tables))
		for _, table := range schema.Tables {
			if !cfg.ShouldIgnoreTable(table.Name) {
				filteredTables = append(filteredTables, table)
			}
		}
		schema.Tables = filteredTables
	}

	slog.Info("Found tables to process", "count", len(schema.Tables))
	for _, table := range schema.Tables {
		slog.Debug("Table details", "name", table.Name, "columns", len(table.Columns))
	}

	// Choose generation strategy based on flags
	if incremental {
		return runIncrementalGeneration(cfg, schema)
	} else if parallel {
		return runParallelGeneration(cfg, schema)
	} else {
		return runStandardGeneration(cfg, schema)
	}
}

// runIncrementalGeneration runs incremental code generation
func runIncrementalGeneration(cfg *config.Config, schema *introspector.Schema) error {
	slog.Info("Using incremental generation")

	// For now, use standard generation
	return runStandardGeneration(cfg, schema)
}

// runParallelGeneration runs parallel code generation
func runParallelGeneration(cfg *config.Config, schema *introspector.Schema) error {
	slog.Info("Using parallel generation", "workers", workers)

	// Enable parallel in config and use standard generation
	cfg.Parallel.Enabled = true
	cfg.Parallel.Workers = workers

	return runStandardGeneration(cfg, schema)
}

// runStandardGeneration runs standard code generation with optional optimizations
func runStandardGeneration(cfg *config.Config, schema *introspector.Schema) error {
	slog.Info("Using standard generation")

	// Create generator
	gen := generator.New(cfg)

	// Generate code
	slog.Info("Generating code...")
	ctx := context.Background()
	if err := gen.Generate(ctx, schema, cfg.OutputDir); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	slog.Info("Code generation completed successfully", "output_dir", cfg.GetBaseDir())
	return nil
}

func setupLogging() {
	var level slog.Level

	if debug {
		level = slog.LevelDebug
	} else if verbose {
		level = slog.LevelInfo
	} else {
		level = slog.LevelWarn
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func loadConfig() (*config.Config, error) {
	cfg := &config.Config{}

	// If no config file specified, try to find one automatically
	if configFile == "" {
		configFile = findDefaultConfigFile()
		if configFile != "" {
			slog.Info("Found configuration file", "file", configFile)
		}
	}

	// Load from config file if specified or found
	if configFile != "" {
		slog.Info("Loading configuration from file", "file", configFile)
		if err := cfg.LoadFromFile(configFile); err != nil {
			return nil, err
		}
		slog.Debug("Schema loaded from config file", "schema", cfg.Schema)
	}

	// Override with command line flags
	if dsn != "" {
		cfg.DSN = dsn
	}
	if schema != "" {
		slog.Debug("Overriding schema from CLI flag", "schema", schema)
		cfg.Schema = schema
	}
	if outputDir != "" {
		cfg.OutputDir = outputDir
	}

	// Override individual output directories if specified via CLI flags
	if modelsDir != "" {
		cfg.OutputDirs.Models = modelsDir
	}
	if interfacesDir != "" {
		cfg.OutputDirs.Interfaces = interfacesDir
	}
	if reposDir != "" {
		cfg.OutputDirs.Repos = reposDir
	}
	if mocksDir != "" {
		cfg.OutputDirs.Mocks = mocksDir
	}
	if testsDir != "" {
		cfg.OutputDirs.Tests = testsDir
	}

	if len(tables) > 0 {
		cfg.Tables = tables
	}
	if templateDir != "" {
		cfg.TemplateDir = templateDir
	}
	if mockProvider != "" {
		cfg.MockProvider = mockProvider
	}
	cfg.WithTests = withTests

	// Apply defaults before validation
	cfg.ApplyDefaults()

	// Validate required fields
	if cfg.DSN == "" {
		return nil, fmt.Errorf("DSN is required (use --dsn flag or config file)")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// findDefaultConfigFile searches for default configuration files in the current directory
func findDefaultConfigFile() string {
	// List of default config file names to search for (in order of preference)
	defaultFiles := []string{
		"pgx-goose-conf.yaml",
		"pgx-goose-conf.yml",
		"pgx-goose-conf.json",
	}

	for _, filename := range defaultFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	return ""
}
