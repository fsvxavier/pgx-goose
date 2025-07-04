package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

func TestRepositoryPostgresTemplate_SQLGeneration(t *testing.T) {
	// Mock table with typical structure
	table := introspector.Table{
		Name: "users",
		Columns: []introspector.Column{
			{Name: "id", GoType: "int", IsPrimaryKey: true},
			{Name: "name", GoType: "string", IsPrimaryKey: false},
			{Name: "email", GoType: "string", IsPrimaryKey: false},
			{Name: "created_at", GoType: "time.Time", IsPrimaryKey: false},
		},
		PrimaryKeys: []string{"id"},
	}

	// Test data for template (matching the actual structure used in generator.go)
	data := struct {
		Table           introspector.Table
		StructName      string
		InterfaceName   string
		ImplName        string
		Package         string
		PrimaryKeyType  string
		PrimaryKeyCol   string
		PrimaryKeyField string
	}{
		Table:           table,
		StructName:      "User",
		InterfaceName:   "UserRepository",
		ImplName:        "UserRepositoryImpl",
		Package:         "postgres",
		PrimaryKeyType:  "int",
		PrimaryKeyCol:   "id",
		PrimaryKeyField: "ID",
	}

	// Get template
	gen := &Generator{}
	tmpl, err := gen.getEmbeddedTemplate("repository_postgres.tmpl")
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	// Execute template
	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	generated := buf.String()

	t.Run("INSERT_SQL_Parameters", func(t *testing.T) {
		// Check INSERT query has correct parameter indexing
		assert.Contains(t, generated, "INSERT INTO users (")
		assert.Contains(t, generated, "name, email, created_at")
		assert.Contains(t, generated, ") VALUES (")
		assert.Contains(t, generated, "$1, $2, $3")
		assert.Contains(t, generated, ") RETURNING id")

		// Extract just the INSERT section to validate
		lines := strings.Split(generated, "\n")
		var insertSection []string
		inInsert := false
		for _, line := range lines {
			if strings.Contains(line, "INSERT INTO") {
				inInsert = true
			}
			if inInsert {
				insertSection = append(insertSection, line)
				if strings.Contains(line, ") RETURNING") {
					break
				}
			}
		}
		insertSQL := strings.Join(insertSection, "\n")

		// Should NOT contain incorrect parameters in INSERT section
		assert.NotContains(t, insertSQL, "$4") // in VALUES clause
		assert.NotContains(t, insertSQL, "$0") // zero index
	})

	t.Run("UPDATE_SQL_Parameters", func(t *testing.T) {
		// Check UPDATE query has correct parameter indexing
		assert.Contains(t, generated, "UPDATE users SET")
		assert.Contains(t, generated, "name = $1")
		assert.Contains(t, generated, "email = $2")
		assert.Contains(t, generated, "created_at = $3")
		assert.Contains(t, generated, "WHERE id = $4")

		// Should NOT contain gaps in parameter indexing
		assert.NotContains(t, generated, "name = $2, email = $4") // skip index
		assert.NotContains(t, generated, "WHERE id = $5")         // wrong final index
	})

	t.Run("Generated_Code_Structure", func(t *testing.T) {
		// Verify the struct and methods are generated correctly
		assert.Contains(t, generated, "type UserRepositoryImpl struct")
		assert.Contains(t, generated, "func NewUserRepository")
		assert.Contains(t, generated, "func (r *UserRepositoryImpl) Create")
		assert.Contains(t, generated, "func (r *UserRepositoryImpl) GetByID")
		assert.Contains(t, generated, "func (r *UserRepositoryImpl) Update")
		assert.Contains(t, generated, "func (r *UserRepositoryImpl) Delete")
		assert.Contains(t, generated, "func (r *UserRepositoryImpl) List")
		assert.Contains(t, generated, "func (r *UserRepositoryImpl) Count")
	})

	t.Run("Parameter_Consistency", func(t *testing.T) {
		// Ensure parameters match the struct fields being passed
		assert.Contains(t, generated, "user.Name,")
		assert.Contains(t, generated, "user.Email,")
		assert.Contains(t, generated, "user.CreatedAt,")

		// In UPDATE, primary key should be last parameter
		lines := strings.Split(generated, "\n")
		var updateSection []string
		inUpdate := false
		for _, line := range lines {
			if strings.Contains(line, "func (r *UserRepositoryImpl) Update") {
				inUpdate = true
			}
			if inUpdate {
				updateSection = append(updateSection, line)
				if strings.Contains(line, "return err") && len(updateSection) > 5 {
					break
				}
			}
		}

		updateCode := strings.Join(updateSection, "\n")
		assert.Contains(t, updateCode, "user.Id,") // PK should be in parameters
	})
}

func TestRepositoryPostgresTemplate_NoLeadingCommas(t *testing.T) {
	// Test with single column (edge case)
	table := introspector.Table{
		Name: "simple",
		Columns: []introspector.Column{
			{Name: "id", GoType: "int", IsPrimaryKey: true},
			{Name: "name", GoType: "string", IsPrimaryKey: false},
		},
		PrimaryKeys: []string{"id"},
	}

	data := struct {
		Table           introspector.Table
		StructName      string
		InterfaceName   string
		ImplName        string
		Package         string
		PrimaryKeyType  string
		PrimaryKeyCol   string
		PrimaryKeyField string
	}{
		Table:           table,
		StructName:      "Simple",
		InterfaceName:   "SimpleRepository",
		ImplName:        "simpleRepository",
		Package:         "postgres",
		PrimaryKeyType:  "int",
		PrimaryKeyCol:   "id",
		PrimaryKeyField: "ID",
	}

	gen := &Generator{}
	tmpl, err := gen.getEmbeddedTemplate("repository_postgres.tmpl")
	require.NoError(t, err)

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	generated := buf.String()

	// Should not have leading commas in INSERT section
	insertLines := []string{}
	lines := strings.Split(generated, "\n")
	inInsert := false
	for _, line := range lines {
		if strings.Contains(line, "INSERT INTO") {
			inInsert = true
		}
		if inInsert {
			insertLines = append(insertLines, line)
			if strings.Contains(line, ") RETURNING") {
				break
			}
		}
	}
	insertSQL := strings.Join(insertLines, "\n")

	// Check for leading commas in INSERT columns and values
	assert.NotContains(t, insertSQL, "(, name") // leading comma in column list
	assert.NotContains(t, insertSQL, "(, $1")   // leading comma in values list

	// Should not have leading commas in UPDATE SET clause
	updateLines := []string{}
	inUpdate := false
	for _, line := range lines {
		if strings.Contains(line, "UPDATE") && strings.Contains(line, "SET") {
			inUpdate = true
		}
		if inUpdate {
			updateLines = append(updateLines, line)
			if strings.Contains(line, "WHERE") {
				break
			}
		}
	}
	updateSQL := strings.Join(updateLines, "\n")
	assert.NotContains(t, updateSQL, "SET, name =") // leading comma in update set
}
