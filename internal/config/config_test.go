package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadFromFile_YAML(t *testing.T) {
	// Create temporary YAML file
	yamlContent := `
dsn: "postgres://test:test@localhost:5432/testdb"
schema: "inventory"
out: "./test-output"
tables: ["users", "orders"]
template_dir: "./templates"
mock_provider: "testify"
with_tests: true
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test loading
	cfg := &Config{}
	err = cfg.LoadFromFile(tmpFile.Name())

	assert.NoError(t, err)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.DSN)
	assert.Equal(t, "inventory", cfg.Schema)
	assert.Equal(t, "./test-output", cfg.OutputDir)
	assert.Equal(t, []string{"users", "orders"}, cfg.Tables)
	assert.Equal(t, "./templates", cfg.TemplateDir)
	assert.Equal(t, "testify", cfg.MockProvider)
	assert.True(t, cfg.WithTests)
}

func TestConfig_LoadFromFile_JSON(t *testing.T) {
	// Create temporary JSON file
	jsonContent := `{
  "dsn": "postgres://test:test@localhost:5432/testdb",
  "schema": "public",
  "out": "./test-output",
  "tables": ["users", "orders"],
  "template_dir": "./templates",
  "mock_provider": "mock",
  "with_tests": false
}`
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test loading
	cfg := &Config{}
	err = cfg.LoadFromFile(tmpFile.Name())

	assert.NoError(t, err)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.DSN)
	assert.Equal(t, "public", cfg.Schema)
	assert.Equal(t, "./test-output", cfg.OutputDir)
	assert.Equal(t, []string{"users", "orders"}, cfg.Tables)
	assert.Equal(t, "./templates", cfg.TemplateDir)
	assert.Equal(t, "mock", cfg.MockProvider)
	assert.False(t, cfg.WithTests)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				DSN:          "postgres://test:test@localhost:5432/testdb",
				MockProvider: "testify",
			},
			wantErr: false,
		},
		{
			name: "missing DSN",
			config: Config{
				MockProvider: "testify",
			},
			wantErr: true,
			errMsg:  "DSN is required",
		},
		{
			name: "invalid mock provider",
			config: Config{
				DSN:          "postgres://test:test@localhost:5432/testdb",
				MockProvider: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid mock provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected Config
	}{
		{
			name: "apply schema default",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
			},
			expected: Config{
				DSN:          "postgres://test:test@localhost:5432/testdb",
				Schema:       "public",
				MockProvider: "testify",
				OutputDirs: OutputDirs{
					Base:       "./pgx-goose",
					Models:     "./pgx-goose/models",
					Interfaces: "./pgx-goose/repository/interfaces",
					Repos:      "./pgx-goose/repository/postgres",
					Mocks:      "./pgx-goose/mocks",
					Tests:      "./pgx-goose/tests",
				},
				OutputDir: "./pgx-goose",
			},
		},
		{
			name: "preserve custom schema",
			config: Config{
				DSN:    "postgres://test:test@localhost:5432/testdb",
				Schema: "inventory",
			},
			expected: Config{
				DSN:          "postgres://test:test@localhost:5432/testdb",
				Schema:       "inventory",
				MockProvider: "testify",
				OutputDirs: OutputDirs{
					Base:       "./pgx-goose",
					Models:     "./pgx-goose/models",
					Interfaces: "./pgx-goose/repository/interfaces",
					Repos:      "./pgx-goose/repository/postgres",
					Mocks:      "./pgx-goose/mocks",
					Tests:      "./pgx-goose/tests",
				},
				OutputDir: "./pgx-goose",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.ApplyDefaults()
			assert.Equal(t, tt.expected.Schema, tt.config.Schema)
			assert.Equal(t, tt.expected.MockProvider, tt.config.MockProvider)
			assert.Equal(t, tt.expected.OutputDir, tt.config.OutputDir)
		})
	}
}

func TestConfig_ShouldIgnoreTable(t *testing.T) {
	tests := []struct {
		name         string
		ignoreTables []string
		tableName    string
		expected     bool
	}{
		{
			name:         "should ignore table in list",
			ignoreTables: []string{"migrations", "logs", "sessions"},
			tableName:    "migrations",
			expected:     true,
		},
		{
			name:         "should ignore table case insensitive",
			ignoreTables: []string{"Migrations", "LOGS"},
			tableName:    "migrations",
			expected:     true,
		},
		{
			name:         "should not ignore table not in list",
			ignoreTables: []string{"migrations", "logs"},
			tableName:    "users",
			expected:     false,
		},
		{
			name:         "should not ignore when list is empty",
			ignoreTables: []string{},
			tableName:    "users",
			expected:     false,
		},
		{
			name:         "should not ignore when list is nil",
			ignoreTables: nil,
			tableName:    "users",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				IgnoreTables: tt.ignoreTables,
			}
			result := cfg.ShouldIgnoreTable(tt.tableName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_FilterTables(t *testing.T) {
	tests := []struct {
		name         string
		ignoreTables []string
		inputTables  []string
		expected     []string
	}{
		{
			name:         "filter out ignored tables",
			ignoreTables: []string{"migrations", "logs"},
			inputTables:  []string{"users", "migrations", "orders", "logs", "products"},
			expected:     []string{"users", "orders", "products"},
		},
		{
			name:         "no filtering when ignore list is empty",
			ignoreTables: []string{},
			inputTables:  []string{"users", "orders", "products"},
			expected:     []string{"users", "orders", "products"},
		},
		{
			name:         "case insensitive filtering",
			ignoreTables: []string{"MIGRATIONS", "logs"},
			inputTables:  []string{"users", "Migrations", "orders", "LOGS"},
			expected:     []string{"users", "orders"},
		},
		{
			name:         "all tables filtered out",
			ignoreTables: []string{"users", "orders"},
			inputTables:  []string{"users", "orders"},
			expected:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				IgnoreTables: tt.ignoreTables,
			}
			result := cfg.FilterTables(tt.inputTables)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_ValidateTableConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		tables       []string
		ignoreTables []string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "valid configuration - no conflicts",
			tables:       []string{"users", "orders"},
			ignoreTables: []string{"migrations", "logs"},
			expectError:  false,
		},
		{
			name:         "valid configuration - empty lists",
			tables:       []string{},
			ignoreTables: []string{},
			expectError:  false,
		},
		{
			name:         "invalid configuration - table in both lists",
			tables:       []string{"users", "orders"},
			ignoreTables: []string{"users", "logs"},
			expectError:  true,
			errorMessage: "table 'users' is specified in both 'tables' and 'ignore_tables' - this is conflicting",
		},
		{
			name:         "invalid configuration - case insensitive conflict",
			tables:       []string{"Users", "orders"},
			ignoreTables: []string{"users", "logs"},
			expectError:  true,
			errorMessage: "table 'Users' is specified in both 'tables' and 'ignore_tables' - this is conflicting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Tables:       tt.tables,
				IgnoreTables: tt.ignoreTables,
			}
			err := cfg.ValidateTableConfiguration()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_LoadFromFile_WithIgnoreTables_YAML(t *testing.T) {
	// Create temporary YAML file with ignore_tables
	yamlContent := `
dsn: "postgres://test:test@localhost:5432/testdb"
schema: "public"
out: "./test-output"
tables: ["users", "orders"]
ignore_tables: ["migrations", "logs", "sessions"]
template_dir: "./templates"
mock_provider: "testify"
with_tests: true
`
	tmpFile, err := os.CreateTemp("", "test-config-ignore-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test loading
	cfg := &Config{}
	err = cfg.LoadFromFile(tmpFile.Name())

	assert.NoError(t, err)
	assert.Equal(t, []string{"users", "orders"}, cfg.Tables)
	assert.Equal(t, []string{"migrations", "logs", "sessions"}, cfg.IgnoreTables)
}

func TestConfig_LoadFromFile_WithIgnoreTables_JSON(t *testing.T) {
	// Create temporary JSON file with ignore_tables
	jsonContent := `{
  "dsn": "postgres://test:test@localhost:5432/testdb",
  "schema": "public",
  "out": "./test-output",
  "tables": ["users", "orders"],
  "ignore_tables": ["migrations", "logs", "sessions"],
  "template_dir": "./templates",
  "mock_provider": "testify",
  "with_tests": true
}`
	tmpFile, err := os.CreateTemp("", "test-config-ignore-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test loading
	cfg := &Config{}
	err = cfg.LoadFromFile(tmpFile.Name())

	assert.NoError(t, err)
	assert.Equal(t, []string{"users", "orders"}, cfg.Tables)
	assert.Equal(t, []string{"migrations", "logs", "sessions"}, cfg.IgnoreTables)
}

func TestConfig_LoadFromFile_SchemaHandling(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectedSchema string
	}{
		{
			name: "load custom schema from YAML",
			configContent: `
dsn: "postgres://test:test@localhost:5432/testdb"
schema: "inventory"
out: "./test-output"
tables: []
ignore_tables: []
`,
			expectedSchema: "inventory",
		},
		{
			name: "load default schema when not specified",
			configContent: `
dsn: "postgres://test:test@localhost:5432/testdb"
out: "./test-output"
tables: []
ignore_tables: []
`,
			expectedSchema: "public", // Should be set by ApplyDefaults()
		},
		{
			name: "load empty schema gets defaulted",
			configContent: `
dsn: "postgres://test:test@localhost:5432/testdb"
schema: ""
out: "./test-output"
tables: []
ignore_tables: []
`,
			expectedSchema: "public", // Should be set by ApplyDefaults()
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "test-schema-config-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.configContent)
			require.NoError(t, err)
			tmpFile.Close()

			// Load configuration
			cfg := &Config{}
			err = cfg.LoadFromFile(tmpFile.Name())
			require.NoError(t, err)

			// Apply defaults (like the real application does)
			cfg.ApplyDefaults()

			// Verify schema
			assert.Equal(t, tt.expectedSchema, cfg.Schema)
		})
	}
}

func TestConfig_LoadFromFile_SchemaJSONHandling(t *testing.T) {
	jsonContent := `{
  "dsn": "postgres://test:test@localhost:5432/testdb",
  "schema": "analytics",
  "out": "./test-output",
  "tables": [],
  "ignore_tables": []
}`

	tmpFile, err := os.CreateTemp("", "test-schema-config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load configuration
	cfg := &Config{}
	err = cfg.LoadFromFile(tmpFile.Name())
	require.NoError(t, err)

	// Apply defaults
	cfg.ApplyDefaults()

	// Verify schema
	assert.Equal(t, "analytics", cfg.Schema)
}

func TestConfig_ApplyDefaults_AdvancedFeatures(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected Config
	}{
		{
			name: "apply advanced feature defaults",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
			},
			expected: Config{
				DSN:          "postgres://test:test@localhost:5432/testdb",
				Schema:       "public",
				MockProvider: "testify",
				OutputDirs: OutputDirs{
					Base:       "./pgx-goose",
					Models:     "./pgx-goose/models",
					Interfaces: "./pgx-goose/repository/interfaces",
					Repos:      "./pgx-goose/repository/postgres",
					Mocks:      "./pgx-goose/mocks",
					Tests:      "./pgx-goose/tests",
				},
				OutputDir: "./pgx-goose",
				Parallel: ParallelConfig{
					Workers: 4,
				},
				TemplateOptimization: TemplateOptimizationConfig{
					CacheSize: 100,
				},
				Migrations: MigrationConfig{
					Format:        "goose",
					NamingPattern: "20060102150405_{{.name}}.sql",
					OutputDir:     "./migrations",
				},
			},
		},
		{
			name: "preserve custom advanced settings",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				Parallel: ParallelConfig{
					Enabled: true,
					Workers: 8,
				},
				TemplateOptimization: TemplateOptimizationConfig{
					Enabled:    true,
					CacheSize:  200,
					Precompile: true,
				},
				CrossSchema: CrossSchemaConfig{
					Enabled: true,
					Schemas: []string{"public", "inventory"},
				},
			},
			expected: Config{
				DSN:          "postgres://test:test@localhost:5432/testdb",
				Schema:       "public",
				MockProvider: "testify",
				OutputDirs: OutputDirs{
					Base:       "./pgx-goose",
					Models:     "./pgx-goose/models",
					Interfaces: "./pgx-goose/repository/interfaces",
					Repos:      "./pgx-goose/repository/postgres",
					Mocks:      "./pgx-goose/mocks",
					Tests:      "./pgx-goose/tests",
				},
				OutputDir: "./pgx-goose",
				Parallel: ParallelConfig{
					Enabled: true,
					Workers: 8,
				},
				TemplateOptimization: TemplateOptimizationConfig{
					Enabled:    true,
					CacheSize:  200,
					Precompile: true,
				},
				CrossSchema: CrossSchemaConfig{
					Enabled: true,
					Schemas: []string{"public", "inventory"},
				},
				Migrations: MigrationConfig{
					Format:        "goose",
					NamingPattern: "20060102150405_{{.name}}.sql",
					OutputDir:     "./migrations",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.ApplyDefaults()
			assert.Equal(t, tt.expected.Parallel.Workers, tt.config.Parallel.Workers)
			assert.Equal(t, tt.expected.TemplateOptimization.CacheSize, tt.config.TemplateOptimization.CacheSize)
			assert.Equal(t, tt.expected.Migrations.Format, tt.config.Migrations.Format)
			assert.Equal(t, tt.expected.Migrations.NamingPattern, tt.config.Migrations.NamingPattern)
		})
	}
}

func TestConfig_ValidateAdvancedFeatures(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid parallel config",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				Parallel: ParallelConfig{
					Enabled: true,
					Workers: 4,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid parallel workers - too low",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				Parallel: ParallelConfig{
					Enabled: true,
					Workers: 0,
				},
			},
			wantErr: true,
			errMsg:  "parallel workers must be at least 1",
		},
		{
			name: "invalid parallel workers - too high",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				Parallel: ParallelConfig{
					Enabled: true,
					Workers: 50,
				},
			},
			wantErr: true,
			errMsg:  "parallel workers cannot exceed 32",
		},
		{
			name: "invalid template cache size",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				TemplateOptimization: TemplateOptimizationConfig{
					Enabled:   true,
					CacheSize: 0,
				},
			},
			wantErr: true,
			errMsg:  "template cache size must be at least 1",
		},
		{
			name: "invalid migration format",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				Migrations: MigrationConfig{
					Enabled: true,
					Format:  "invalid",
				},
			},
			wantErr: true,
			errMsg:  "unsupported migration format",
		},
		{
			name: "missing migration output dir",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				Migrations: MigrationConfig{
					Enabled: true,
					Format:  "goose",
				},
			},
			wantErr: true,
			errMsg:  "migration output directory is required",
		},
		{
			name: "cross-schema enabled without schemas",
			config: Config{
				DSN: "postgres://test:test@localhost:5432/testdb",
				CrossSchema: CrossSchemaConfig{
					Enabled: true,
					Schemas: []string{},
				},
			},
			wantErr: true,
			errMsg:  "at least one schema must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				tt.config.ApplyDefaults()
			}
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_AdvancedFeatureHelpers(t *testing.T) {
	cfg := &Config{
		Parallel: ParallelConfig{
			Enabled: true,
		},
		TemplateOptimization: TemplateOptimizationConfig{
			Enabled: true,
		},
		Incremental: IncrementalConfig{
			Enabled: true,
		},
		CrossSchema: CrossSchemaConfig{
			Enabled: true,
		},
		Migrations: MigrationConfig{
			Enabled: true,
		},
		GoGenerate: GoGenerateConfig{
			Enabled: true,
		},
	}

	assert.True(t, cfg.IsParallelEnabled())
	assert.True(t, cfg.IsTemplateOptimizationEnabled())
	assert.True(t, cfg.IsIncrementalEnabled())
	assert.True(t, cfg.IsCrossSchemaEnabled())
	assert.True(t, cfg.IsMigrationsEnabled())
	assert.True(t, cfg.IsGoGenerateEnabled())
}

func TestConfig_LoadFromFile_WithAdvancedFeatures_YAML(t *testing.T) {
	yamlContent := `
dsn: "postgres://test:test@localhost:5432/testdb"
schema: "public"
out: "./test-output"
parallel:
  enabled: true
  workers: 8
template_optimization:
  enabled: true
  cache_size: 200
  precompile: true
incremental:
  enabled: true
  force: false
cross_schema:
  enabled: true
  schemas: ["public", "inventory"]
  relationship_detection: true
migrations:
  enabled: true
  output_dir: "./migrations"
  format: "goose"
  naming_pattern: "20060102150405_{{.name}}.sql"
go_generate:
  enabled: true
  create_directive: true
  update_makefile: true
`
	tmpFile, err := os.CreateTemp("", "test-advanced-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg := &Config{}
	err = cfg.LoadFromFile(tmpFile.Name())

	assert.NoError(t, err)
	assert.True(t, cfg.Parallel.Enabled)
	assert.Equal(t, 8, cfg.Parallel.Workers)
	assert.True(t, cfg.TemplateOptimization.Enabled)
	assert.Equal(t, 200, cfg.TemplateOptimization.CacheSize)
	assert.True(t, cfg.TemplateOptimization.Precompile)
	assert.True(t, cfg.Incremental.Enabled)
	assert.False(t, cfg.Incremental.Force)
	assert.True(t, cfg.CrossSchema.Enabled)
	assert.Equal(t, []string{"public", "inventory"}, cfg.CrossSchema.Schemas)
	assert.True(t, cfg.CrossSchema.RelationshipDetection)
	assert.True(t, cfg.Migrations.Enabled)
	assert.Equal(t, "./migrations", cfg.Migrations.OutputDir)
	assert.Equal(t, "goose", cfg.Migrations.Format)
	assert.Equal(t, "20060102150405_{{.name}}.sql", cfg.Migrations.NamingPattern)
	assert.True(t, cfg.GoGenerate.Enabled)
	assert.True(t, cfg.GoGenerate.CreateDirective)
	assert.True(t, cfg.GoGenerate.UpdateMakefile)
}
