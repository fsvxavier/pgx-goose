package generator

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMigrationGenerator(t *testing.T) {
	cfg := &config.Config{
		OutputDir: "./test-output",
	}
	cfg.ApplyDefaults()

	mg := NewMigrationGenerator(cfg)

	assert.NotNil(t, mg)
	assert.NotNil(t, mg.config)
	assert.NotNil(t, mg.optimizer)
	assert.Contains(t, mg.migrationDir, "migrations")
}

func TestMigrationGenerator_CalculateSchemaDiff_NewSchema(t *testing.T) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	// No old schema (first migration)
	newSchema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
					{Name: "name", Type: "varchar", IsNullable: false},
				},
			},
		},
	}

	diff, err := mg.calculateSchemaDiff(nil, newSchema)
	require.NoError(t, err)

	assert.Len(t, diff.AddedTables, 1)
	assert.Equal(t, "users", diff.AddedTables[0].Name)
	assert.Len(t, diff.DroppedTables, 0)
	assert.Len(t, diff.ModifiedTables, 0)
}

func TestMigrationGenerator_CalculateSchemaDiff_ModifiedSchema(t *testing.T) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	oldSchema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
					{Name: "name", Type: "varchar", IsNullable: false},
				},
				Indexes: []introspector.Index{
					{Name: "idx_name", Columns: []string{"name"}},
				},
			},
			{
				Name: "products",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
				},
			},
		},
	}

	newSchema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
					{Name: "name", Type: "varchar", IsNullable: false},
					{Name: "email", Type: "varchar", IsNullable: true}, // Added column
				},
				Indexes: []introspector.Index{
					{Name: "idx_name", Columns: []string{"name"}},
					{Name: "idx_email", Columns: []string{"email"}}, // Added index
				},
				ForeignKeys: []introspector.ForeignKey{
					{Name: "fk_user_profile", Column: "profile_id", ReferencedTable: "profiles", ReferencedColumn: "id"}, // Added FK
				},
			},
			{
				Name: "categories", // New table
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
				},
			},
			// products table removed
		},
	}

	diff, err := mg.calculateSchemaDiff(oldSchema, newSchema)
	require.NoError(t, err)

	// Check added tables
	assert.Len(t, diff.AddedTables, 1)
	assert.Equal(t, "categories", diff.AddedTables[0].Name)

	// Check dropped tables
	assert.Len(t, diff.DroppedTables, 1)
	assert.Equal(t, "products", diff.DroppedTables[0])

	// Check added columns
	assert.Contains(t, diff.AddedColumns, "users")
	assert.Len(t, diff.AddedColumns["users"], 1)
	assert.Equal(t, "email", diff.AddedColumns["users"][0].Name)

	// Check added indexes
	assert.Contains(t, diff.AddedIndexes, "users")
	assert.Len(t, diff.AddedIndexes["users"], 1)
	assert.Equal(t, "idx_email", diff.AddedIndexes["users"][0].Name)

	// Check added foreign keys
	assert.Contains(t, diff.AddedForeignKeys, "users")
	assert.Len(t, diff.AddedForeignKeys["users"], 1)
	assert.Equal(t, "fk_user_profile", diff.AddedForeignKeys["users"][0].Name)
}

func TestMigrationGenerator_CompareColumn(t *testing.T) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	tests := []struct {
		name       string
		oldCol     introspector.Column
		newCol     introspector.Column
		expectDiff bool
		changeType ColumnChangeType
	}{
		{
			name: "no changes",
			oldCol: introspector.Column{
				Name: "test_col", Type: "varchar", IsNullable: true, DefaultValue: nil,
			},
			newCol: introspector.Column{
				Name: "test_col", Type: "varchar", IsNullable: true, DefaultValue: nil,
			},
			expectDiff: false,
		},
		{
			name: "type changed",
			oldCol: introspector.Column{
				Name: "test_col", Type: "varchar", IsNullable: true,
			},
			newCol: introspector.Column{
				Name: "test_col", Type: "text", IsNullable: true,
			},
			expectDiff: true,
			changeType: ColumnTypeChanged,
		},
		{
			name: "nullability changed",
			oldCol: introspector.Column{
				Name: "test_col", Type: "varchar", IsNullable: true,
			},
			newCol: introspector.Column{
				Name: "test_col", Type: "varchar", IsNullable: false,
			},
			expectDiff: true,
			changeType: ColumnNullabilityChanged,
		},
		{
			name: "default value changed",
			oldCol: introspector.Column{
				Name: "test_col", Type: "varchar", IsNullable: true, DefaultValue: stringPtr("old_default"),
			},
			newCol: introspector.Column{
				Name: "test_col", Type: "varchar", IsNullable: true, DefaultValue: stringPtr("new_default"),
			},
			expectDiff: true,
			changeType: ColumnDefaultChanged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := mg.compareColumn(tt.oldCol, tt.newCol)

			if tt.expectDiff {
				require.NotNil(t, diff)
				assert.Equal(t, tt.changeType, diff.ChangeType)
				assert.Equal(t, tt.oldCol.Type, diff.OldType)
				assert.Equal(t, tt.newCol.Type, diff.NewType)
			} else {
				assert.Nil(t, diff)
			}
		})
	}
}

func TestMigrationGenerator_IsDiffEmpty(t *testing.T) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	tests := []struct {
		name     string
		diff     *SchemaDiff
		expected bool
	}{
		{
			name: "empty diff",
			diff: &SchemaDiff{
				AddedColumns:       make(map[string][]introspector.Column),
				DroppedColumns:     make(map[string][]string),
				ModifiedColumns:    make(map[string][]ColumnDiff),
				AddedIndexes:       make(map[string][]introspector.Index),
				DroppedIndexes:     make(map[string][]string),
				AddedForeignKeys:   make(map[string][]introspector.ForeignKey),
				DroppedForeignKeys: make(map[string][]string),
			},
			expected: true,
		},
		{
			name: "has added tables",
			diff: &SchemaDiff{
				AddedTables:        []introspector.Table{{Name: "test"}},
				AddedColumns:       make(map[string][]introspector.Column),
				DroppedColumns:     make(map[string][]string),
				ModifiedColumns:    make(map[string][]ColumnDiff),
				AddedIndexes:       make(map[string][]introspector.Index),
				DroppedIndexes:     make(map[string][]string),
				AddedForeignKeys:   make(map[string][]introspector.ForeignKey),
				DroppedForeignKeys: make(map[string][]string),
			},
			expected: false,
		},
		{
			name: "has added columns",
			diff: &SchemaDiff{
				AddedColumns: map[string][]introspector.Column{
					"users": {{Name: "email"}},
				},
				DroppedColumns:     make(map[string][]string),
				ModifiedColumns:    make(map[string][]ColumnDiff),
				AddedIndexes:       make(map[string][]introspector.Index),
				DroppedIndexes:     make(map[string][]string),
				AddedForeignKeys:   make(map[string][]introspector.ForeignKey),
				DroppedForeignKeys: make(map[string][]string),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mg.isDiffEmpty(tt.diff)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMigrationGenerator_GenerateCreateTableSQL(t *testing.T) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	tables := []introspector.Table{
		{
			Name: "users",
			Columns: []introspector.Column{
				{Name: "id", Type: "SERIAL", IsNullable: false, IsPrimaryKey: true},
				{Name: "name", Type: "VARCHAR(255)", IsNullable: false},
				{Name: "email", Type: "VARCHAR(255)", IsNullable: true, DefaultValue: stringPtr("NULL")},
			},
			PrimaryKeys: []string{"id"},
		},
	}

	sql, err := mg.generateCreateTableSQL(tables)
	require.NoError(t, err)

	assert.Contains(t, sql, "CREATE TABLE users")
	assert.Contains(t, sql, "id SERIAL NOT NULL")
	assert.Contains(t, sql, "name VARCHAR(255) NOT NULL")
	assert.Contains(t, sql, "email VARCHAR(255) DEFAULT NULL")
	assert.Contains(t, sql, "PRIMARY KEY (id)")
}

func TestMigrationGenerator_GenerateDropTableSQL(t *testing.T) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	tables := []introspector.Table{
		{Name: "users"},
		{Name: "products"},
		{Name: "orders"},
	}

	sql := mg.generateDropTableSQL(tables)

	// Should drop in reverse order
	lines := strings.Split(strings.TrimSpace(sql), "\n")
	assert.Contains(t, lines[0], "DROP TABLE IF EXISTS orders")
	assert.Contains(t, lines[1], "DROP TABLE IF EXISTS products")
	assert.Contains(t, lines[2], "DROP TABLE IF EXISTS users")
}

func TestMigrationGenerator_WriteGooseMigration(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	mg := NewMigrationGenerator(cfg)
	mg.migrationDir = tempDir

	migration := Migration{
		Version:     "20250107120000",
		Name:        "20250107120000_create_users",
		UpSQL:       "CREATE TABLE users (id SERIAL PRIMARY KEY);",
		DownSQL:     "DROP TABLE users;",
		Description: "Create users table",
		Timestamp:   time.Now(),
	}

	err := mg.writeGooseMigration(migration)
	require.NoError(t, err)

	// Check file was created
	expectedFile := fmt.Sprintf("%s/20250107120000_20250107120000_create_users.sql", tempDir)
	assert.FileExists(t, expectedFile)

	// Check file content
	content, err := os.ReadFile(expectedFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "-- +goose Up")
	assert.Contains(t, contentStr, "-- +goose Down")
	assert.Contains(t, contentStr, "CREATE TABLE users")
	assert.Contains(t, contentStr, "DROP TABLE users")
}

func TestMigrationGenerator_EqualStringPointers(t *testing.T) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	tests := []struct {
		name     string
		a        *string
		b        *string
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "first nil",
			a:        nil,
			b:        stringPtr("test"),
			expected: false,
		},
		{
			name:     "second nil",
			a:        stringPtr("test"),
			b:        nil,
			expected: false,
		},
		{
			name:     "both equal",
			a:        stringPtr("test"),
			b:        stringPtr("test"),
			expected: true,
		},
		{
			name:     "both different",
			a:        stringPtr("test1"),
			b:        stringPtr("test2"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mg.equalStringPointers(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmarks

func BenchmarkMigrationGenerator_CalculateSchemaDiff(b *testing.B) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	// Create large schemas for benchmarking
	oldTables := make([]introspector.Table, 100)
	newTables := make([]introspector.Table, 100)

	for i := 0; i < 100; i++ {
		oldTables[i] = introspector.Table{
			Name: fmt.Sprintf("table_%d", i),
			Columns: []introspector.Column{
				{Name: "id", Type: "int", IsPrimaryKey: true},
				{Name: "name", Type: "varchar", IsNullable: false},
			},
			Indexes: []introspector.Index{
				{Name: fmt.Sprintf("idx_%d", i), Columns: []string{"name"}},
			},
		}

		// New schema has slight modifications
		newTables[i] = introspector.Table{
			Name: fmt.Sprintf("table_%d", i),
			Columns: []introspector.Column{
				{Name: "id", Type: "int", IsPrimaryKey: true},
				{Name: "name", Type: "varchar", IsNullable: false},
				{Name: "created_at", Type: "timestamp", IsNullable: true}, // Added column
			},
			Indexes: []introspector.Index{
				{Name: fmt.Sprintf("idx_%d", i), Columns: []string{"name"}},
			},
		}
	}

	oldSchema := &introspector.Schema{Tables: oldTables}
	newSchema := &introspector.Schema{Tables: newTables}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := mg.calculateSchemaDiff(oldSchema, newSchema)
		require.NoError(b, err)
	}
}

func BenchmarkMigrationGenerator_GenerateCreateTableSQL(b *testing.B) {
	cfg := &config.Config{}
	mg := NewMigrationGenerator(cfg)

	// Create tables for benchmarking
	tables := make([]introspector.Table, 50)
	for i := 0; i < 50; i++ {
		columns := make([]introspector.Column, 20)
		for j := 0; j < 20; j++ {
			columns[j] = introspector.Column{
				Name:         fmt.Sprintf("column_%d", j),
				Type:         "varchar",
				IsNullable:   j%2 == 0,
				IsPrimaryKey: j == 0,
			}
		}

		tables[i] = introspector.Table{
			Name:        fmt.Sprintf("table_%d", i),
			Columns:     columns,
			PrimaryKeys: []string{"column_0"},
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := mg.generateCreateTableSQL(tables)
		require.NoError(b, err)
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}
