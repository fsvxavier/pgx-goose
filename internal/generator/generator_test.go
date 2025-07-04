package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_profile", "UserProfile"},
		{"user_profile_settings", "UserProfileSettings"},
		{"", ""},
		{"single", "Single"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := toPascalCase(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"UserProfileSettings", "user_profile_settings"},
		{"", ""},
		{"Single", "single"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := toSnakeCase(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetPrimaryKeyType(t *testing.T) {
	g := &Generator{}

	table := introspector.Table{
		Name: "users",
		Columns: []introspector.Column{
			{Name: "id", GoType: "int64", IsPrimaryKey: true},
			{Name: "name", GoType: "string", IsPrimaryKey: false},
		},
	}

	result := g.getPrimaryKeyType(table)
	assert.Equal(t, "int64", result)

	// Test with no primary key
	tableNoPK := introspector.Table{
		Name: "logs",
		Columns: []introspector.Column{
			{Name: "message", GoType: "string", IsPrimaryKey: false},
		},
	}

	result = g.getPrimaryKeyType(tableNoPK)
	assert.Equal(t, "interface{}", result)
}

func TestGetPrimaryKeyColumn(t *testing.T) {
	g := &Generator{}

	table := introspector.Table{
		Name: "users",
		Columns: []introspector.Column{
			{Name: "id", GoType: "int64", IsPrimaryKey: true},
			{Name: "name", GoType: "string", IsPrimaryKey: false},
		},
	}

	result := g.getPrimaryKeyColumn(table)
	assert.Equal(t, "id", result)

	// Test with no primary key
	tableNoPK := introspector.Table{
		Name: "logs",
		Columns: []introspector.Column{
			{Name: "message", GoType: "string", IsPrimaryKey: false},
		},
	}

	result = g.getPrimaryKeyColumn(tableNoPK)
	assert.Equal(t, "id", result)
}
