package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// OutputDirs represents the output directories for different types of generated files.
type OutputDirs struct {
	Base       string `json:"base"         yaml:"base"`         // Base output directory
	Models     string `json:"models"       yaml:"models"`       // Models output directory
	Interfaces string `json:"interfaces"   yaml:"interfaces"`   // Repository interfaces directory
	Repos      string `json:"repositories" yaml:"repositories"` // Repository implementations directory
	Mocks      string `json:"mocks"        yaml:"mocks"`        // Mocks directory
	Tests      string `json:"tests"        yaml:"tests"`        // Tests directory
}

// ParallelConfig represents parallel generation configuration.
type ParallelConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"` // Enable parallel generation
	Workers int  `json:"workers" yaml:"workers"` // Number of worker goroutines
}

// TemplateOptimizationConfig represents template optimization configuration.
type TemplateOptimizationConfig struct {
	CacheSize  int  `json:"cache_size" yaml:"cache_size"`
	Enabled    bool `json:"enabled"    yaml:"enabled"`
	Precompile bool `json:"precompile" yaml:"precompile"`
}

// IncrementalConfig represents incremental generation configuration.
type IncrementalConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"` // Enable incremental generation
	Force   bool `json:"force"   yaml:"force"`   // Force full regeneration
}

// CrossSchemaConfig represents cross-schema configuration.
type CrossSchemaConfig struct {
	Schemas               []string `json:"schemas"                yaml:"schemas"`
	Enabled               bool     `json:"enabled"                yaml:"enabled"`
	RelationshipDetection bool     `json:"relationship_detection" yaml:"relationship_detection"`
}

// MigrationConfig represents migration generation configuration.
type MigrationConfig struct {
	OutputDir     string `json:"output_dir"     yaml:"output_dir"`
	Format        string `json:"format"         yaml:"format"`
	NamingPattern string `json:"naming_pattern" yaml:"naming_pattern"`
	Enabled       bool   `json:"enabled"        yaml:"enabled"`
}

// GoGenerateConfig represents go:generate integration configuration.
type GoGenerateConfig struct {
	Enabled           bool `json:"enabled"             yaml:"enabled"`             // Enable go:generate integration
	CreateDirective   bool `json:"create_directive"    yaml:"create_directive"`    // Create go:generate directive
	UpdateMakefile    bool `json:"update_makefile"     yaml:"update_makefile"`     // Update Makefile
	UpdateVSCodeTasks bool `json:"update_vscode_tasks" yaml:"update_vscode_tasks"` // Update VS Code tasks
	UpdateGitignore   bool `json:"update_gitignore"    yaml:"update_gitignore"`    // Update .gitignore
}

// Config represents the configuration for pgx-goose.
type Config struct {
	OutputDirs           OutputDirs                 `json:"output_dirs"   yaml:"output_dirs"`
	Migrations           MigrationConfig            `json:"migrations"            yaml:"migrations"`
	TemplateDir          string                     `json:"template_dir"  yaml:"template_dir"`
	OutputDir            string                     `json:"out"           yaml:"out"`
	DSN                  string                     `json:"dsn"           yaml:"dsn"`
	MockProvider         string                     `json:"mock_provider" yaml:"mock_provider"`
	Schema               string                     `json:"schema"        yaml:"schema"`
	Tables               []string                   `json:"tables"        yaml:"tables"`
	IgnoreTables         []string                   `json:"ignore_tables" yaml:"ignore_tables"`
	CrossSchema          CrossSchemaConfig          `json:"cross_schema"          yaml:"cross_schema"`
	TemplateOptimization TemplateOptimizationConfig `json:"template_optimization" yaml:"template_optimization"`
	Parallel             ParallelConfig             `json:"parallel"              yaml:"parallel"`
	GoGenerate           GoGenerateConfig           `json:"go_generate"           yaml:"go_generate"`
	Incremental          IncrementalConfig          `json:"incremental"           yaml:"incremental"`
	WithTests            bool                       `json:"with_tests"    yaml:"with_tests"`
}

// LoadFromFile loads configuration from a YAML or JSON file.
func (c *Config) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, c)
	case ".json":
		return json.Unmarshal(data, c)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// SaveToFile saves configuration to a YAML or JSON file.
func (c *Config) SaveToFile(filename string) error {
	var data []byte
	var err error

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c)
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filename, data, 0o644)
}

// ApplyDefaults applies default values to the configuration.
func (c *Config) ApplyDefaults() {
	// Set default schema
	if c.Schema == "" {
		c.Schema = "public"
	}

	// Set default mock provider
	if c.MockProvider == "" {
		c.MockProvider = "testify"
	}

	// Set default output directories based on legacy OutputDir or defaults
	baseDir := c.OutputDir
	if baseDir == "" && c.OutputDirs.Base == "" {
		baseDir = "./pgx-goose"
	} else if c.OutputDirs.Base != "" {
		baseDir = c.OutputDirs.Base
	}

	// Apply defaults to OutputDirs if not specified
	if c.OutputDirs.Base == "" {
		c.OutputDirs.Base = baseDir
	}
	if c.OutputDirs.Models == "" {
		c.OutputDirs.Models = filepath.Join(baseDir, "models")
	}
	if c.OutputDirs.Interfaces == "" {
		c.OutputDirs.Interfaces = filepath.Join(baseDir, "repository", "interfaces")
	}
	if c.OutputDirs.Repos == "" {
		c.OutputDirs.Repos = filepath.Join(baseDir, "repository", "postgres")
	}
	if c.OutputDirs.Mocks == "" {
		c.OutputDirs.Mocks = filepath.Join(baseDir, "mocks")
	}
	if c.OutputDirs.Tests == "" {
		c.OutputDirs.Tests = filepath.Join(baseDir, "tests")
	}

	// Ensure legacy OutputDir is in sync with OutputDirs.Base
	if c.OutputDir == "" {
		c.OutputDir = c.OutputDirs.Base
	}

	// Apply defaults for advanced features
	c.applyAdvancedDefaults()
}

// applyAdvancedDefaults applies default values for advanced features.
func (c *Config) applyAdvancedDefaults() {
	// Parallel configuration defaults
	if c.Parallel.Workers == 0 {
		c.Parallel.Workers = 4 // Default to 4 workers
	}

	// Template optimization defaults
	if c.TemplateOptimization.CacheSize == 0 {
		c.TemplateOptimization.CacheSize = 100 // Default cache size
	}

	// Cross-schema defaults
	if len(c.CrossSchema.Schemas) == 0 && c.CrossSchema.Enabled {
		c.CrossSchema.Schemas = []string{"public"} // Default to public schema
	}

	// Migration defaults
	if c.Migrations.Format == "" {
		c.Migrations.Format = "goose"
	}
	if c.Migrations.NamingPattern == "" {
		c.Migrations.NamingPattern = "20060102150405_{{.name}}.sql"
	}
	if c.Migrations.OutputDir == "" {
		c.Migrations.OutputDir = "./migrations"
	}
}

// GetModelsDir returns the models output directory.
func (c *Config) GetModelsDir() string {
	if c.OutputDirs.Models != "" {
		return c.OutputDirs.Models
	}
	return filepath.Join(c.GetBaseDir(), "models")
}

// GetInterfacesDir returns the interfaces output directory.
func (c *Config) GetInterfacesDir() string {
	if c.OutputDirs.Interfaces != "" {
		return c.OutputDirs.Interfaces
	}
	return filepath.Join(c.GetBaseDir(), "repository", "interfaces")
}

// GetReposDir returns the repository implementations output directory.
func (c *Config) GetReposDir() string {
	if c.OutputDirs.Repos != "" {
		return c.OutputDirs.Repos
	}
	return filepath.Join(c.GetBaseDir(), "repository", "postgres")
}

// GetMocksDir returns the mocks output directory.
func (c *Config) GetMocksDir() string {
	if c.OutputDirs.Mocks != "" {
		return c.OutputDirs.Mocks
	}
	return filepath.Join(c.GetBaseDir(), "mocks")
}

// GetTestsDir returns the tests output directory.
func (c *Config) GetTestsDir() string {
	if c.OutputDirs.Tests != "" {
		return c.OutputDirs.Tests
	}
	return filepath.Join(c.GetBaseDir(), "tests")
}

// GetBaseDir returns the base output directory.
func (c *Config) GetBaseDir() string {
	if c.OutputDirs.Base != "" {
		return c.OutputDirs.Base
	}
	if c.OutputDir != "" {
		return c.OutputDir
	}
	return "./pgx-goose"
}

// GetAllOutputDirs returns all output directories.
func (c *Config) GetAllOutputDirs() []string {
	dirs := []string{
		c.GetModelsDir(),
		c.GetInterfacesDir(),
		c.GetReposDir(),
		c.GetMocksDir(),
	}

	if c.WithTests {
		dirs = append(dirs, c.GetTestsDir())
	}

	return dirs
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("DSN is required")
	}

	if c.MockProvider != "" && c.MockProvider != "testify" && c.MockProvider != "mock" {
		return fmt.Errorf("invalid mock provider: %s (must be 'testify' or 'mock')", c.MockProvider)
	}

	// Validate table configuration
	if err := c.ValidateTableConfiguration(); err != nil {
		return err
	}

	// Validate advanced features
	if err := c.validateAdvancedFeatures(); err != nil {
		return err
	}

	return nil
}

// validateAdvancedFeatures validates advanced feature configurations.
func (c *Config) validateAdvancedFeatures() error {
	// Validate parallel configuration - only if enabled or workers explicitly set
	if c.Parallel.Enabled || c.Parallel.Workers > 0 {
		if c.Parallel.Workers < 1 {
			return fmt.Errorf("parallel workers must be at least 1, got: %d", c.Parallel.Workers)
		}
		if c.Parallel.Workers > 32 {
			return fmt.Errorf("parallel workers cannot exceed 32, got: %d", c.Parallel.Workers)
		}
	}

	// Validate template optimization configuration - only if enabled or cache size explicitly set
	if c.TemplateOptimization.Enabled || c.TemplateOptimization.CacheSize > 0 {
		if c.TemplateOptimization.CacheSize < 1 {
			return fmt.Errorf("template cache size must be at least 1, got: %d", c.TemplateOptimization.CacheSize)
		}
		if c.TemplateOptimization.CacheSize > 1000 {
			return fmt.Errorf("template cache size cannot exceed 1000, got: %d", c.TemplateOptimization.CacheSize)
		}
	}

	// Validate migration configuration
	if c.Migrations.Enabled {
		if c.Migrations.Format != "goose" {
			return fmt.Errorf("unsupported migration format: %s (currently only 'goose' is supported)", c.Migrations.Format)
		}
		if c.Migrations.OutputDir == "" {
			return fmt.Errorf("migration output directory is required when migrations are enabled")
		}
	}

	// Validate cross-schema configuration
	if c.CrossSchema.Enabled && len(c.CrossSchema.Schemas) == 0 {
		return fmt.Errorf("at least one schema must be specified when cross-schema is enabled")
	}

	return nil
}

// GetMigrationsDir returns the migrations output directory.
func (c *Config) GetMigrationsDir() string {
	if c.Migrations.OutputDir != "" {
		return c.Migrations.OutputDir
	}
	return "./migrations"
}

// IsParallelEnabled returns true if parallel generation is enabled.
func (c *Config) IsParallelEnabled() bool {
	return c.Parallel.Enabled
}

// IsIncrementalEnabled returns true if incremental generation is enabled.
func (c *Config) IsIncrementalEnabled() bool {
	return c.Incremental.Enabled
}

// IsCrossSchemaEnabled returns true if cross-schema support is enabled.
func (c *Config) IsCrossSchemaEnabled() bool {
	return c.CrossSchema.Enabled
}

// IsMigrationsEnabled returns true if migration generation is enabled.
func (c *Config) IsMigrationsEnabled() bool {
	return c.Migrations.Enabled
}

// IsGoGenerateEnabled returns true if go:generate integration is enabled.
func (c *Config) IsGoGenerateEnabled() bool {
	return c.GoGenerate.Enabled
}

// IsTemplateOptimizationEnabled returns true if template optimization is enabled.
func (c *Config) IsTemplateOptimizationEnabled() bool {
	return c.TemplateOptimization.Enabled
}

// ShouldIgnoreTable checks if a table should be ignored.
func (c *Config) ShouldIgnoreTable(tableName string) bool {
	for _, ignoredTable := range c.IgnoreTables {
		if strings.EqualFold(ignoredTable, tableName) {
			return true
		}
	}
	return false
}

// FilterTables filters a list of tables, removing ignored ones.
func (c *Config) FilterTables(tables []string) []string {
	if len(c.IgnoreTables) == 0 {
		return tables
	}

	filtered := make([]string, 0, len(tables))
	for _, table := range tables {
		if !c.ShouldIgnoreTable(table) {
			filtered = append(filtered, table)
		}
	}
	return filtered
}

// ValidateTableConfiguration validates table and ignore_tables configuration.
func (c *Config) ValidateTableConfiguration() error {
	// Check for conflicts between tables and ignore_tables
	for _, table := range c.Tables {
		if c.ShouldIgnoreTable(table) {
			return fmt.Errorf("table '%s' is specified in both 'tables' and 'ignore_tables' - this is conflicting", table)
		}
	}
	return nil
}
