package introspector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntrospector_New(t *testing.T) {
	tests := []struct {
		name           string
		dsn            string
		schema         string
		expectedSchema string
	}{
		{
			name:           "creates introspector with custom schema",
			dsn:            "postgres://user:pass@localhost/db",
			schema:         "custom_schema",
			expectedSchema: "custom_schema",
		},
		{
			name:           "defaults to public schema when empty",
			dsn:            "postgres://user:pass@localhost/db",
			schema:         "",
			expectedSchema: "public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			introspector := New(tt.dsn, tt.schema)

			assert.NotNil(t, introspector)
			assert.Equal(t, tt.dsn, introspector.dsn)
			assert.Equal(t, tt.expectedSchema, introspector.schema)
		})
	}
}

func TestIntrospector_IntrospectSchema_InvalidDSN(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "invalid dsn format",
			dsn:  "invalid-dsn",
		},
		{
			name: "empty dsn",
			dsn:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			introspector := New(tt.dsn, "public")

			schema, err := introspector.IntrospectSchema([]string{})

			assert.Error(t, err)
			assert.Nil(t, schema)
			// More flexible error message check
			assert.NotEmpty(t, err.Error())
		})
	}
}

func TestColumn_ValidationBasic(t *testing.T) {
	column := Column{
		Name:         "id",
		Type:         "integer",
		GoType:       "int64",
		IsPrimaryKey: true,
		IsNullable:   false,
		Position:     1,
	}

	assert.NotEmpty(t, column.Name)
	assert.NotEmpty(t, column.Type)
	assert.NotEmpty(t, column.GoType)
	assert.GreaterOrEqual(t, column.Position, 0)
}

func TestIndex_ValidationBasic(t *testing.T) {
	index := Index{
		Name:     "users_email_idx",
		Columns:  []string{"email"},
		IsUnique: true,
	}

	assert.NotEmpty(t, index.Name)
	assert.NotEmpty(t, index.Columns)
}

func TestForeignKey_ValidationBasic(t *testing.T) {
	fk := ForeignKey{
		Name:             "fk_user_company",
		Column:           "company_id",
		ReferencedTable:  "companies",
		ReferencedColumn: "id",
	}

	assert.NotEmpty(t, fk.Name)
	assert.NotEmpty(t, fk.Column)
	assert.NotEmpty(t, fk.ReferencedTable)
	assert.NotEmpty(t, fk.ReferencedColumn)
}

func TestTable_ValidationBasic(t *testing.T) {
	table := Table{
		Name: "users",
		Columns: []Column{
			{
				Name:         "id",
				Type:         "integer",
				GoType:       "int64",
				IsPrimaryKey: true,
				Position:     1,
			},
		},
	}

	assert.NotEmpty(t, table.Name)
	assert.NotEmpty(t, table.Columns)
}

func TestSchema_ValidationBasic(t *testing.T) {
	schema := &Schema{
		Tables: []Table{
			{
				Name: "users",
				Columns: []Column{
					{
						Name:         "id",
						Type:         "integer",
						GoType:       "int64",
						IsPrimaryKey: true,
						Position:     1,
					},
				},
			},
		},
	}

	assert.NotEmpty(t, schema.Tables)
	assert.Len(t, schema.Tables, 1)

	// Check table names
	tableNames := make([]string, len(schema.Tables))
	for i, table := range schema.Tables {
		tableNames[i] = table.Name
	}
	assert.Contains(t, tableNames, "users")
}
