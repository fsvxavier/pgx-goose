package generator

import (
	"fmt"
	"testing"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCrossSchemaGenerator(t *testing.T) {
	cfg := &config.Config{
		OutputDir: "./test-output",
	}

	csg := NewCrossSchemaGenerator(cfg)

	assert.NotNil(t, csg)
	assert.NotNil(t, csg.Generator)
	assert.NotNil(t, csg.schemas)
	assert.NotNil(t, csg.crossReferences)
}

func TestCrossSchemaGenerator_ParseCrossSchemaReference(t *testing.T) {
	cfg := &config.Config{}
	csg := NewCrossSchemaGenerator(cfg)

	// Add test schemas
	csg.schemas["public"] = &introspector.Schema{}
	csg.schemas["auth"] = &introspector.Schema{}

	tests := []struct {
		name           string
		sourceSchema   string
		sourceTable    string
		fk             introspector.ForeignKey
		expectedResult *CrossReference
	}{
		{
			name:         "cross-schema reference",
			sourceSchema: "public",
			sourceTable:  "orders",
			fk: introspector.ForeignKey{
				Name:             "fk_order_user",
				Column:           "user_id",
				ReferencedTable:  "auth.users",
				ReferencedColumn: "id",
			},
			expectedResult: &CrossReference{
				SourceSchema:   "public",
				SourceTable:    "orders",
				SourceColumn:   "user_id",
				TargetSchema:   "auth",
				TargetTable:    "users",
				TargetColumn:   "id",
				RelationType:   ManyToOne,
				ForeignKeyName: "fk_order_user",
			},
		},
		{
			name:         "same schema reference",
			sourceSchema: "public",
			sourceTable:  "orders",
			fk: introspector.ForeignKey{
				Name:             "fk_order_product",
				Column:           "product_id",
				ReferencedTable:  "public.products",
				ReferencedColumn: "id",
			},
			expectedResult: nil, // Same schema, should return nil
		},
		{
			name:         "non-existent target schema",
			sourceSchema: "public",
			sourceTable:  "orders",
			fk: introspector.ForeignKey{
				Name:             "fk_order_external",
				Column:           "external_id",
				ReferencedTable:  "nonexistent.external",
				ReferencedColumn: "id",
			},
			expectedResult: nil, // Target schema doesn't exist
		},
		{
			name:         "simple table reference",
			sourceSchema: "public",
			sourceTable:  "orders",
			fk: introspector.ForeignKey{
				Name:             "fk_order_simple",
				Column:           "simple_id",
				ReferencedTable:  "simple_table",
				ReferencedColumn: "id",
			},
			expectedResult: nil, // Not cross-schema format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csg.parseCrossSchemaReference(tt.sourceSchema, tt.sourceTable, tt.fk)

			if tt.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.SourceSchema, result.SourceSchema)
				assert.Equal(t, tt.expectedResult.SourceTable, result.SourceTable)
				assert.Equal(t, tt.expectedResult.SourceColumn, result.SourceColumn)
				assert.Equal(t, tt.expectedResult.TargetSchema, result.TargetSchema)
				assert.Equal(t, tt.expectedResult.TargetTable, result.TargetTable)
				assert.Equal(t, tt.expectedResult.TargetColumn, result.TargetColumn)
				assert.Equal(t, tt.expectedResult.RelationType, result.RelationType)
				assert.Equal(t, tt.expectedResult.ForeignKeyName, result.ForeignKeyName)
			}
		})
	}
}

func TestCrossSchemaGenerator_FilterTables(t *testing.T) {
	tests := []struct {
		name         string
		tables       []string
		ignoreTables []string
		expected     []string
	}{
		{
			name:         "no ignore tables",
			tables:       []string{"users", "products", "orders"},
			ignoreTables: []string{},
			expected:     []string{"users", "products", "orders"},
		},
		{
			name:         "with ignore tables",
			tables:       []string{"users", "products", "orders", "migrations"},
			ignoreTables: []string{"migrations", "temp_table"},
			expected:     []string{"users", "products", "orders"},
		},
		{
			name:         "case insensitive filtering",
			tables:       []string{"Users", "Products", "ORDERS"},
			ignoreTables: []string{"users", "PRODUCTS"},
			expected:     []string{"ORDERS"},
		},
		{
			name:         "all tables ignored",
			tables:       []string{"migrations", "temp_table"},
			ignoreTables: []string{"migrations", "temp_table"},
			expected:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				IgnoreTables: tt.ignoreTables,
			}
			result := cfg.FilterTables(tt.tables)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCrossSchemaGenerator_FilterIgnoredTables(t *testing.T) {
	cfg := &config.Config{}
	csg := NewCrossSchemaGenerator(cfg)

	tables := []introspector.Table{
		{Name: "users"},
		{Name: "products"},
		{Name: "migrations"},
		{Name: "temp_table"},
	}

	tests := []struct {
		name          string
		ignoreTables  []string
		expectedLen   int
		expectedNames []string
	}{
		{
			name:          "no ignore tables",
			ignoreTables:  []string{},
			expectedLen:   4,
			expectedNames: []string{"users", "products", "migrations", "temp_table"},
		},
		{
			name:          "ignore some tables",
			ignoreTables:  []string{"migrations", "temp_table"},
			expectedLen:   2,
			expectedNames: []string{"users", "products"},
		},
		{
			name:          "case insensitive",
			ignoreTables:  []string{"USERS", "products"},
			expectedLen:   2,
			expectedNames: []string{"migrations", "temp_table"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csg.filterIgnoredTables(tables, tt.ignoreTables)
			assert.Len(t, result, tt.expectedLen)

			var names []string
			for _, table := range result {
				names = append(names, table.Name)
			}
			assert.ElementsMatch(t, tt.expectedNames, names)
		})
	}
}

func TestCrossSchemaGenerator_AddCrossReference(t *testing.T) {
	cfg := &config.Config{}
	csg := NewCrossSchemaGenerator(cfg)

	ref1 := CrossReference{
		SourceSchema: "public",
		SourceTable:  "orders",
		TargetSchema: "auth",
		TargetTable:  "users",
	}

	ref2 := CrossReference{
		SourceSchema: "public",
		SourceTable:  "products",
		TargetSchema: "inventory",
		TargetTable:  "stock",
	}

	// Add references
	csg.addCrossReference(ref1)
	csg.addCrossReference(ref2)

	// Verify references are stored
	publicRefs := csg.GetCrossReferences("public")
	assert.Len(t, publicRefs, 2)

	// Verify specific references
	assert.Contains(t, publicRefs, ref1)
	assert.Contains(t, publicRefs, ref2)

	// Verify no references for other schemas
	authRefs := csg.GetCrossReferences("auth")
	assert.Len(t, authRefs, 0)
}

func TestCrossSchemaGenerator_GetAllSchemas(t *testing.T) {
	cfg := &config.Config{}
	csg := NewCrossSchemaGenerator(cfg)

	// Add test schemas
	schema1 := &introspector.Schema{
		Tables: []introspector.Table{{Name: "users"}},
	}
	schema2 := &introspector.Schema{
		Tables: []introspector.Table{{Name: "products"}},
	}

	csg.schemas["public"] = schema1
	csg.schemas["inventory"] = schema2

	allSchemas := csg.GetAllSchemas()

	assert.Len(t, allSchemas, 2)
	assert.Contains(t, allSchemas, "public")
	assert.Contains(t, allSchemas, "inventory")
	assert.Equal(t, schema1, allSchemas["public"])
	assert.Equal(t, schema2, allSchemas["inventory"])
}

// Benchmarks

func BenchmarkCrossSchemaGenerator_ParseCrossSchemaReference(b *testing.B) {
	cfg := &config.Config{}
	csg := NewCrossSchemaGenerator(cfg)

	// Setup schemas
	for i := 0; i < 10; i++ {
		schemaName := fmt.Sprintf("schema_%d", i)
		csg.schemas[schemaName] = &introspector.Schema{}
	}

	fk := introspector.ForeignKey{
		Name:             "fk_test",
		Column:           "ref_id",
		ReferencedTable:  "schema_5.target_table",
		ReferencedColumn: "id",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = csg.parseCrossSchemaReference("schema_1", "source_table", fk)
	}
}

func BenchmarkCrossSchemaGenerator_FilterTables(b *testing.B) {
	cfg := &config.Config{}
	csg := NewCrossSchemaGenerator(cfg)

	// Create large lists for benchmarking
	tables := make([]string, 1000)
	ignoreTables := make([]string, 100)

	for i := 0; i < 1000; i++ {
		tables[i] = fmt.Sprintf("table_%d", i)
	}

	for i := 0; i < 100; i++ {
		ignoreTables[i] = fmt.Sprintf("table_%d", i*10) // Every 10th table
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = csg.filterTables(tables, ignoreTables)
	}
}

func BenchmarkCrossSchemaGenerator_FilterIgnoredTables(b *testing.B) {
	cfg := &config.Config{}
	csg := NewCrossSchemaGenerator(cfg)

	// Create large table list
	tables := make([]introspector.Table, 1000)
	for i := 0; i < 1000; i++ {
		tables[i] = introspector.Table{Name: fmt.Sprintf("table_%d", i)}
	}

	ignoreTables := make([]string, 100)
	for i := 0; i < 100; i++ {
		ignoreTables[i] = fmt.Sprintf("table_%d", i*10)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = csg.filterIgnoredTables(tables, ignoreTables)
	}
}
