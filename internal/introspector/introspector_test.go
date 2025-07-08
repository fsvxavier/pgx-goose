package introspector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapPostgresToGoType(t *testing.T) {
	tests := []struct {
		pgType     string
		isNullable bool
		expected   string
	}{
		{"integer", false, "int32"},
		{"integer", true, "*int32"},
		{"bigint", false, "int64"},
		{"bigint", true, "*int64"},
		{"text", false, "string"},
		{"text", true, "*string"},
		{"boolean", false, "bool"},
		{"boolean", true, "*bool"},
		{"timestamp", false, "time.Time"},
		{"timestamp", true, "*time.Time"},
		{"uuid", false, "string"},
		{"uuid", true, "*string"},
		{"json", false, "json.RawMessage"},
		{"json", true, "json.RawMessage"},
		{"unknown_type", false, "interface{}"},
		{"unknown_type", true, "interface{}"},
	}

	for _, test := range tests {
		t.Run(test.pgType, func(t *testing.T) {
			result := mapPostgresToGoType(test.pgType, test.isNullable)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestNewIntrospector(t *testing.T) {
	dsn := "postgres://test:test@localhost:5432/testdb"

	// Test with default schema
	introspector1 := New(dsn, "")
	assert.NotNil(t, introspector1)
	assert.Equal(t, dsn, introspector1.dsn)
	assert.Equal(t, "public", introspector1.schema)

	// Test with custom schema
	introspector2 := New(dsn, "inventory")
	assert.NotNil(t, introspector2)
	assert.Equal(t, dsn, introspector2.dsn)
	assert.Equal(t, "inventory", introspector2.schema)
}
