package introspector

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockPoolAdapter implements a mock for database connection pool.
type MockPoolAdapter struct {
	pingErr                   error
	queryErr                  error
	queryResult               *MockRowsResult
	queryRowResult            *MockRowResult
	multiQueryResults         []*MockRowsResult
	queryIndex                int
	closed                    bool
	introspectTableShouldFail bool
}

func (m *MockPoolAdapter) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *MockPoolAdapter) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}

	// Handle multiple query results for complex tests
	if m.multiQueryResults != nil && m.queryIndex < len(m.multiQueryResults) {
		result := m.multiQueryResults[m.queryIndex]
		m.queryIndex++
		return result, nil
	}

	return m.queryResult, nil
}

func (m *MockPoolAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if m.queryRowResult != nil {
		return m.queryRowResult
	}
	return &MockRowResult{}
}

func (m *MockPoolAdapter) Close() {
	m.closed = true
}

// MockRowsResult implements pgx.Rows interface for testing.
type MockRowsResult struct {
	err        error
	rows       [][]interface{}
	currentRow int
}

func (m *MockRowsResult) Next() bool {
	m.currentRow++
	return m.currentRow < len(m.rows)
}

func (m *MockRowsResult) Scan(dest ...interface{}) error {
	if m.currentRow >= len(m.rows) {
		return fmt.Errorf("no more rows")
	}

	row := m.rows[m.currentRow]
	for i, value := range row {
		if i < len(dest) {
			switch d := dest[i].(type) {
			case *string:
				if str, ok := value.(string); ok {
					*d = str
				}
			case *int:
				if num, ok := value.(int); ok {
					*d = num
				}
			case *bool:
				if b, ok := value.(bool); ok {
					*d = b
				}
			case *[]string:
				if slice, ok := value.([]string); ok {
					*d = slice
				}
			case **string:
				if value != nil {
					if str, ok := value.(string); ok {
						*d = &str
					}
				} else {
					*d = nil
				}
			}
		}
	}

	return nil
}

func (m *MockRowsResult) Close() {
	// Nothing to do for mock
}

func (m *MockRowsResult) Err() error {
	return m.err
}

func (m *MockRowsResult) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *MockRowsResult) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *MockRowsResult) RawValues() [][]byte {
	return nil
}

func (m *MockRowsResult) Values() ([]interface{}, error) {
	if m.currentRow >= len(m.rows) {
		return nil, fmt.Errorf("no more rows")
	}
	return m.rows[m.currentRow], nil
}

func (m *MockRowsResult) Conn() *pgx.Conn {
	return nil // Mock doesn't need real connection
}

// MockRowResult implements pgx.Row interface for testing.
type MockRowResult struct {
	value interface{}
	err   error
}

func (m *MockRowResult) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}

	if len(dest) > 0 {
		switch d := dest[0].(type) {
		case *string:
			if str, ok := m.value.(string); ok {
				*d = str
			}
		case **string:
			if m.value != nil {
				if str, ok := m.value.(string); ok {
					*d = &str
				}
			} else {
				*d = nil
			}
		}
	}

	return nil
}

// Additional mock implementations for better interface compliance

// MockIntrospectorInterface implements a mock introspector for testing.
type MockIntrospectorInterface struct {
	schema *Schema
	err    error
}

func (m *MockIntrospectorInterface) IntrospectSchema(ctx context.Context, tables []string) (*Schema, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.schema, nil
}

func (m *MockIntrospectorInterface) GetAllTables(ctx context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}

	var tableNames []string
	if m.schema != nil {
		for _, table := range m.schema.Tables {
			tableNames = append(tableNames, table.Name)
		}
	}
	return tableNames, nil
}

func (m *MockIntrospectorInterface) Close() error {
	return m.err
}

// Helper functions for creating test data

func createTestSchema() *Schema {
	return &Schema{
		Tables: []Table{
			{
				Name:    "users",
				Comment: "User management table",
				Columns: []Column{
					{
						Name:         "id",
						Type:         "bigint",
						GoType:       "int64",
						IsPrimaryKey: true,
						IsNullable:   false,
						Position:     1,
						Comment:      "Primary key",
					},
					{
						Name:         "name",
						Type:         "varchar",
						GoType:       "string",
						IsPrimaryKey: false,
						IsNullable:   false,
						Position:     2,
						Comment:      "User name",
					},
					{
						Name:         "email",
						Type:         "varchar",
						GoType:       "*string",
						IsPrimaryKey: false,
						IsNullable:   true,
						Position:     3,
						Comment:      "User email",
					},
				},
				PrimaryKeys: []string{"id"},
				Indexes: []Index{
					{
						Name:     "idx_users_email",
						Columns:  []string{"email"},
						IsUnique: true,
					},
				},
				ForeignKeys: []ForeignKey{},
			},
			{
				Name:    "products",
				Comment: "Product catalog",
				Columns: []Column{
					{
						Name:         "id",
						Type:         "bigint",
						GoType:       "int64",
						IsPrimaryKey: true,
						IsNullable:   false,
						Position:     1,
						Comment:      "Primary key",
					},
					{
						Name:         "name",
						Type:         "varchar",
						GoType:       "string",
						IsPrimaryKey: false,
						IsNullable:   false,
						Position:     2,
						Comment:      "Product name",
					},
					{
						Name:         "price",
						Type:         "decimal",
						GoType:       "float64",
						IsPrimaryKey: false,
						IsNullable:   false,
						Position:     3,
						Comment:      "Product price",
					},
				},
				PrimaryKeys: []string{"id"},
				Indexes:     []Index{},
				ForeignKeys: []ForeignKey{},
			},
		},
	}
}

func createTestTable() Table {
	return Table{
		Name:    "test_table",
		Comment: "Test table for unit tests",
		Columns: []Column{
			{
				Name:         "id",
				Type:         "bigint",
				GoType:       "int64",
				IsPrimaryKey: true,
				IsNullable:   false,
				Position:     1,
				Comment:      "Primary key",
			},
			{
				Name:         "name",
				Type:         "varchar",
				GoType:       "string",
				IsPrimaryKey: false,
				IsNullable:   false,
				Position:     2,
				Comment:      "Name field",
			},
		},
		PrimaryKeys: []string{"id"},
		Indexes:     []Index{},
		ForeignKeys: []ForeignKey{},
	}
}
