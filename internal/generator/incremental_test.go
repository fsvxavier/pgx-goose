package generator

import (
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIncrementalGenerator(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		OutputDir: tempDir,
	}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	assert.NotNil(t, ig)
	assert.NotNil(t, ig.Generator)
	assert.NotNil(t, ig.metadata)
	assert.Contains(t, ig.metadataFile, ".pgx-goose-metadata.json")
}

func TestIncrementalGenerator_CalculateSchemaHash(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	schema1 := &introspector.Schema{
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

	schema2 := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
					{Name: "name", Type: "varchar", IsNullable: false},
					{Name: "email", Type: "varchar", IsNullable: true}, // Added column
				},
			},
		},
	}

	hash1, err := ig.calculateSchemaHash(schema1)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	hash2, err := ig.calculateSchemaHash(schema2)
	require.NoError(t, err)
	assert.NotEmpty(t, hash2)

	// Hashes should be different
	assert.NotEqual(t, hash1, hash2)

	// Same schema should produce same hash
	hash1Again, err := ig.calculateSchemaHash(schema1)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash1Again)
}

func TestIncrementalGenerator_CalculateTableHash(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	table1 := introspector.Table{
		Name: "users",
		Columns: []introspector.Column{
			{Name: "id", Type: "int", IsPrimaryKey: true},
			{Name: "name", Type: "varchar", IsNullable: false},
		},
		ForeignKeys: []introspector.ForeignKey{
			{Name: "fk_user_profile", Column: "profile_id", ReferencedTable: "profiles", ReferencedColumn: "id"},
		},
	}

	table2 := introspector.Table{
		Name: "users",
		Columns: []introspector.Column{
			{Name: "id", Type: "int", IsPrimaryKey: true},
			{Name: "name", Type: "varchar", IsNullable: false},
			{Name: "email", Type: "varchar", IsNullable: true}, // Added column
		},
		ForeignKeys: []introspector.ForeignKey{
			{Name: "fk_user_profile", Column: "profile_id", ReferencedTable: "profiles", ReferencedColumn: "id"},
		},
	}

	hash1 := ig.calculateTableHash(table1)
	hash2 := ig.calculateTableHash(table2)

	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash2)
	assert.NotEqual(t, hash1, hash2)

	// Same table should produce same hash
	hash1Again := ig.calculateTableHash(table1)
	assert.Equal(t, hash1, hash1Again)
}

func TestIncrementalGenerator_DetectChanges(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	// Calculate the actual config hash for consistency
	configHash, err := ig.calculateConfigHash()
	require.NoError(t, err)

	// Define the current schema
	currentSchema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
					{Name: "name", Type: "varchar", IsNullable: false},
					{Name: "email", Type: "varchar", IsNullable: true}, // Modified
				},
			},
			{
				Name: "products",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
					{Name: "title", Type: "varchar", IsNullable: false},
				},
			},
			{
				Name: "categories", // New table
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
				},
			},
			// orders table removed
		},
	}

	// Calculate the actual hash for products table (unchanged)
	productsHash := ig.calculateTableHash(currentSchema.Tables[1])

	// Calculate hash for users table without email column (old version)
	oldUsersTable := introspector.Table{
		Name: "users",
		Columns: []introspector.Column{
			{Name: "id", Type: "int", IsPrimaryKey: true},
			{Name: "name", Type: "varchar", IsNullable: false},
		},
	}
	oldUsersHash := ig.calculateTableHash(oldUsersTable)

	// Setup initial metadata
	ig.metadata.SchemaHash = "old_hash"
	ig.metadata.ConfigHash = configHash // Use actual config hash
	ig.metadata.TableHashes = map[string]string{
		"users":    oldUsersHash, // Use actual old hash
		"products": productsHash, // Use actual current hash (unchanged)
		"orders":   "order_hash", // This will be removed
	}

	changes, err := ig.detectChanges(currentSchema)
	require.NoError(t, err)

	// Should detect:
	// - users table modified
	// - categories table added
	// - orders table removed
	assert.Len(t, changes, 3)

	var addedTables, modifiedTables, removedTables []TableChange
	for _, change := range changes {
		switch change.ChangeType {
		case TableAdded:
			addedTables = append(addedTables, change)
		case TableModified:
			modifiedTables = append(modifiedTables, change)
		case TableRemoved:
			removedTables = append(removedTables, change)
		}
	}

	assert.Len(t, addedTables, 1)
	assert.Equal(t, "categories", addedTables[0].TableName)

	assert.Len(t, modifiedTables, 1)
	assert.Equal(t, "users", modifiedTables[0].TableName)

	assert.Len(t, removedTables, 1)
	assert.Equal(t, "orders", removedTables[0].TableName)
}

func TestIncrementalGenerator_DetectChanges_FirstGeneration(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	// Empty metadata (first generation)
	ig.metadata.SchemaHash = ""
	ig.metadata.TableHashes = make(map[string]string)

	schema := &introspector.Schema{
		Tables: []introspector.Table{
			{Name: "users"},
			{Name: "products"},
		},
	}

	changes, err := ig.detectChanges(schema)
	require.NoError(t, err)

	// All tables should be marked as added
	assert.Len(t, changes, 2)
	for _, change := range changes {
		assert.Equal(t, TableAdded, change.ChangeType)
	}
}

func TestIncrementalGenerator_DetectChanges_NoChanges(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	schema := &introspector.Schema{
		Tables: []introspector.Table{
			{
				Name: "users",
				Columns: []introspector.Column{
					{Name: "id", Type: "int", IsPrimaryKey: true},
				},
			},
		},
	}

	// Set metadata to match current schema
	schemaHash, err := ig.calculateSchemaHash(schema)
	require.NoError(t, err)

	configHash, err := ig.calculateConfigHash()
	require.NoError(t, err)

	ig.metadata.SchemaHash = schemaHash
	ig.metadata.ConfigHash = configHash
	ig.metadata.TableHashes = map[string]string{
		"users": ig.calculateTableHash(schema.Tables[0]),
	}

	changes, err := ig.detectChanges(schema)
	require.NoError(t, err)

	// No changes should be detected
	assert.Len(t, changes, 0)
}

func TestIncrementalGenerator_LoadAndSaveMetadata(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	// Create test metadata
	testMetadata := &GenerationMetadata{
		LastGeneration: time.Now(),
		SchemaHash:     "test_schema_hash",
		ConfigHash:     "test_config_hash",
		TableHashes: map[string]string{
			"users":    "user_hash",
			"products": "product_hash",
		},
		FileHashes: map[string]string{
			"models/user.go":    "file_hash_1",
			"models/product.go": "file_hash_2",
		},
		GeneratedFiles: map[string]GeneratedFileInfo{
			"models/user.go": {
				Path:           "models/user.go",
				Hash:           "file_hash_1",
				Size:           1024,
				ModTime:        time.Now(),
				TableName:      "users",
				GenerationType: "model",
			},
		},
		Version: "1.0",
	}

	ig.metadata = testMetadata

	// Save metadata
	err := ig.saveMetadata()
	require.NoError(t, err)

	// Verify file exists
	assert.FileExists(t, ig.metadataFile)

	// Create new generator and load metadata
	ig2 := NewIncrementalGenerator(cfg)
	err = ig2.loadMetadata()
	require.NoError(t, err)

	// Verify metadata was loaded correctly
	assert.Equal(t, testMetadata.SchemaHash, ig2.metadata.SchemaHash)
	assert.Equal(t, testMetadata.ConfigHash, ig2.metadata.ConfigHash)
	assert.Equal(t, testMetadata.TableHashes, ig2.metadata.TableHashes)
	assert.Equal(t, testMetadata.FileHashes, ig2.metadata.FileHashes)
	assert.Equal(t, len(testMetadata.GeneratedFiles), len(ig2.metadata.GeneratedFiles))
}

func TestIncrementalGenerator_LoadMetadata_FileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	// Load metadata when file doesn't exist
	err := ig.loadMetadata()
	require.NoError(t, err)

	// Should have default metadata
	assert.NotNil(t, ig.metadata)
	assert.Equal(t, "1.0", ig.metadata.Version)
	assert.NotNil(t, ig.metadata.TableHashes)
	assert.NotNil(t, ig.metadata.FileHashes)
	assert.NotNil(t, ig.metadata.GeneratedFiles)
}

func TestIncrementalGenerator_ForceRegeneration(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	// Create metadata file
	ig.metadata.SchemaHash = "test_hash"
	err := ig.saveMetadata()
	require.NoError(t, err)
	assert.FileExists(t, ig.metadataFile)

	// Force regeneration
	err = ig.ForceRegeneration()
	require.NoError(t, err)

	// Metadata file should be removed
	assert.NoFileExists(t, ig.metadataFile)

	// Metadata should be reset
	assert.Empty(t, ig.metadata.SchemaHash)
	assert.Empty(t, ig.metadata.TableHashes)
}

func TestIncrementalGenerator_GetChangedTables(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	schema := &introspector.Schema{
		Tables: []introspector.Table{
			{Name: "users"},
			{Name: "products"},
			{Name: "orders"},
		},
	}

	changes := []TableChange{
		{TableName: "users", ChangeType: TableModified},
		{TableName: "categories", ChangeType: TableAdded}, // Not in schema
		{TableName: "orders", ChangeType: TableRemoved},
	}

	changedTables := ig.getChangedTables(schema, changes)

	// Should only include tables that exist in schema and are added/modified
	assert.Len(t, changedTables, 1)
	assert.Equal(t, "users", changedTables[0].Name)
}

// Benchmarks

func BenchmarkIncrementalGenerator_CalculateSchemaHash(b *testing.B) {
	tempDir := b.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	// Create a large schema for benchmarking
	tables := make([]introspector.Table, 100)
	for i := 0; i < 100; i++ {
		tables[i] = introspector.Table{
			Name: fmt.Sprintf("table_%d", i),
			Columns: []introspector.Column{
				{Name: "id", Type: "int", IsPrimaryKey: true},
				{Name: "name", Type: "varchar", IsNullable: false},
				{Name: "created_at", Type: "timestamp", IsNullable: true},
			},
			ForeignKeys: []introspector.ForeignKey{
				{Name: fmt.Sprintf("fk_%d", i), Column: "ref_id", ReferencedTable: "ref_table", ReferencedColumn: "id"},
			},
		}
	}

	schema := &introspector.Schema{Tables: tables}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ig.calculateSchemaHash(schema)
		require.NoError(b, err)
	}
}

func BenchmarkIncrementalGenerator_CalculateTableHash(b *testing.B) {
	tempDir := b.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	table := introspector.Table{
		Name:        "complex_table",
		Columns:     make([]introspector.Column, 50),
		ForeignKeys: make([]introspector.ForeignKey, 10),
	}

	// Fill with test data
	for i := 0; i < 50; i++ {
		table.Columns[i] = introspector.Column{
			Name:         fmt.Sprintf("column_%d", i),
			Type:         "varchar",
			IsPrimaryKey: i == 0,
			IsNullable:   i%2 == 0,
		}
	}

	for i := 0; i < 10; i++ {
		table.ForeignKeys[i] = introspector.ForeignKey{
			Name:             fmt.Sprintf("fk_%d", i),
			Column:           fmt.Sprintf("ref_id_%d", i),
			ReferencedTable:  "ref_table",
			ReferencedColumn: "id",
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ig.calculateTableHash(table)
	}
}

func BenchmarkIncrementalGenerator_DetectChanges(b *testing.B) {
	tempDir := b.TempDir()
	cfg := &config.Config{OutputDir: tempDir}
	cfg.ApplyDefaults()

	ig := NewIncrementalGenerator(cfg)

	// Setup metadata with many tables
	ig.metadata.SchemaHash = "old_hash"
	ig.metadata.ConfigHash = "config_hash"
	ig.metadata.TableHashes = make(map[string]string)

	for i := 0; i < 1000; i++ {
		ig.metadata.TableHashes[fmt.Sprintf("table_%d", i)] = fmt.Sprintf("hash_%d", i)
	}

	// Create current schema with some changes
	tables := make([]introspector.Table, 1000)
	for i := 0; i < 1000; i++ {
		tables[i] = introspector.Table{
			Name: fmt.Sprintf("table_%d", i),
			Columns: []introspector.Column{
				{Name: "id", Type: "int", IsPrimaryKey: true},
			},
		}
	}

	schema := &introspector.Schema{Tables: tables}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ig.detectChanges(schema)
		require.NoError(b, err)
	}
}
