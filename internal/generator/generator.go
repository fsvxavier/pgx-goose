package generator

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// GeneratorInterface defines the interface for generators
type GeneratorInterface interface {
	Generate(tables []introspector.Table, outputPath string) error
}

// Generator handles code generation
type Generator struct {
	config *config.Config
}

// New creates a new Generator
func New(cfg *config.Config) *Generator {
	return &Generator{config: cfg}
}

// getTemplate loads a template from external file or embedded templates
func (g *Generator) getTemplate(name string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"toPascalCase": toPascalCase,
		"lower":        strings.ToLower,
		"add": func(a, b int) int {
			return a + b
		},
		"slice": func(s string, start, end int) string {
			if start >= len(s) {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	}

	// Try to load from custom template directory first
	if g.config.TemplateDir != "" {
		templatePath := filepath.Join(g.config.TemplateDir, name)
		if _, err := os.Stat(templatePath); err == nil {
			slog.Debug("Loading template from file", "path", templatePath)
			return template.New(name).Funcs(funcMap).ParseFiles(templatePath)
		}
	}

	// Fallback to embedded templates
	slog.Debug("Loading embedded template", "name", name)
	return g.getEmbeddedTemplate(name)
}

// Generate generates all code files
func (g *Generator) Generate(schema *introspector.Schema) error {
	// Create output directory structure
	if err := g.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate models
	if err := g.generateModels(schema); err != nil {
		return fmt.Errorf("failed to generate models: %w", err)
	}

	// Generate repository interfaces
	if err := g.generateRepositoryInterfaces(schema); err != nil {
		return fmt.Errorf("failed to generate repository interfaces: %w", err)
	}

	// Generate repository implementations
	if err := g.generateRepositoryImplementations(schema); err != nil {
		return fmt.Errorf("failed to generate repository implementations: %w", err)
	}

	// Generate mocks
	if err := g.generateMocks(schema); err != nil {
		return fmt.Errorf("failed to generate mocks: %w", err)
	}

	// Generate tests if requested
	if g.config.WithTests {
		if err := g.generateTests(schema); err != nil {
			return fmt.Errorf("failed to generate tests: %w", err)
		}
	}

	return nil
}

// createDirectories creates the necessary directory structure
func (g *Generator) createDirectories() error {
	dirs := g.config.GetAllOutputDirs()

	// Also ensure base directory exists
	dirs = append([]string{g.config.GetBaseDir()}, dirs...)

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// generateModels generates model structs
func (g *Generator) generateModels(schema *introspector.Schema) error {
	slog.Info("Generating models...")

	tmpl, err := g.getTemplate("model.tmpl")
	if err != nil {
		return err
	}

	for _, table := range schema.Tables {
		data := struct {
			Table      introspector.Table
			StructName string
			Package    string
		}{
			Table:      table,
			StructName: toPascalCase(table.Name),
			Package:    "models",
		}

		filename := fmt.Sprintf("%s.go", toSnakeCase(table.Name))
		filepath := filepath.Join(g.config.GetModelsDir(), filename)

		if err := g.writeTemplate(tmpl, filepath, data); err != nil {
			return fmt.Errorf("failed to generate model for table %s: %w", table.Name, err)
		}

		slog.Debug("Generated model", "filename", filename)
	}

	return nil
}

// generateRepositoryInterfaces generates repository interfaces
func (g *Generator) generateRepositoryInterfaces(schema *introspector.Schema) error {
	slog.Info("Generating repository interfaces...")

	tmpl, err := g.getTemplate("repository_interface.tmpl")
	if err != nil {
		return err
	}

	for _, table := range schema.Tables {
		data := struct {
			Table          introspector.Table
			StructName     string
			InterfaceName  string
			Package        string
			PrimaryKeyType string
		}{
			Table:          table,
			StructName:     toPascalCase(table.Name),
			InterfaceName:  toPascalCase(table.Name) + "Repository",
			Package:        "interfaces",
			PrimaryKeyType: g.getPrimaryKeyType(table),
		}

		filename := fmt.Sprintf("%s_repository.go", toSnakeCase(table.Name))
		filepath := filepath.Join(g.config.GetInterfacesDir(), filename)

		if err := g.writeTemplate(tmpl, filepath, data); err != nil {
			return fmt.Errorf("failed to generate repository interface for table %s: %w", table.Name, err)
		}

		slog.Debug("Generated repository interface", "filename", filename)
	}

	return nil
}

// generateRepositoryImplementations generates repository implementations
func (g *Generator) generateRepositoryImplementations(schema *introspector.Schema) error {
	slog.Info("Generating repository implementations...")

	tmpl, err := g.getTemplate("repository_postgres.tmpl")
	if err != nil {
		return err
	}

	for _, table := range schema.Tables {
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
			StructName:      toPascalCase(table.Name),
			InterfaceName:   toPascalCase(table.Name) + "Repository",
			ImplName:        toPascalCase(table.Name) + "Repository",
			Package:         "postgres",
			PrimaryKeyType:  g.getPrimaryKeyType(table),
			PrimaryKeyCol:   g.getPrimaryKeyColumn(table),
			PrimaryKeyField: toPascalCase(g.getPrimaryKeyColumn(table)),
		}

		filename := fmt.Sprintf("%s_repository.go", toSnakeCase(table.Name))
		filepath := filepath.Join(g.config.GetReposDir(), filename)

		if err := g.writeTemplate(tmpl, filepath, data); err != nil {
			return fmt.Errorf("failed to generate repository implementation for table %s: %w", table.Name, err)
		}

		slog.Debug("Generated repository implementation", "filename", filename)
	}

	return nil
}

// generateMocks generates mock implementations
func (g *Generator) generateMocks(schema *introspector.Schema) error {
	slog.Info("Generating mocks...")

	var tmpl *template.Template
	var err error

	switch g.config.MockProvider {
	case "testify":
		tmpl, err = g.getTemplate("mock_testify.tmpl")
	case "mock":
		tmpl, err = g.getTemplate("mock_gomock.tmpl")
	default:
		return fmt.Errorf("unsupported mock provider: %s", g.config.MockProvider)
	}

	if err != nil {
		return err
	}

	for _, table := range schema.Tables {
		data := struct {
			Table          introspector.Table
			StructName     string
			InterfaceName  string
			MockName       string
			Package        string
			PrimaryKeyType string
		}{
			Table:          table,
			StructName:     toPascalCase(table.Name),
			InterfaceName:  toPascalCase(table.Name) + "Repository",
			MockName:       "Mock" + toPascalCase(table.Name) + "Repository",
			Package:        "mocks",
			PrimaryKeyType: g.getPrimaryKeyType(table),
		}

		filename := fmt.Sprintf("mock_%s_repository.go", toSnakeCase(table.Name))
		filepath := filepath.Join(g.config.GetMocksDir(), filename)

		if err := g.writeTemplate(tmpl, filepath, data); err != nil {
			return fmt.Errorf("failed to generate mock for table %s: %w", table.Name, err)
		}

		slog.Debug("Generated mock", "filename", filename)
	}

	return nil
}

// generateTests generates unit tests
func (g *Generator) generateTests(schema *introspector.Schema) error {
	slog.Info("Generating tests...")

	tmpl, err := g.getTemplate("test.tmpl")
	if err != nil {
		return err
	}

	for _, table := range schema.Tables {
		data := struct {
			Table          introspector.Table
			StructName     string
			InterfaceName  string
			MockName       string
			Package        string
			PrimaryKeyType string
		}{
			Table:          table,
			StructName:     toPascalCase(table.Name),
			InterfaceName:  toPascalCase(table.Name) + "Repository",
			MockName:       "Mock" + toPascalCase(table.Name) + "Repository",
			Package:        "tests",
			PrimaryKeyType: g.getPrimaryKeyType(table),
		}

		filename := fmt.Sprintf("%s_repository_test.go", toSnakeCase(table.Name))
		filepath := filepath.Join(g.config.GetTestsDir(), filename)

		if err := g.writeTemplate(tmpl, filepath, data); err != nil {
			return fmt.Errorf("failed to generate test for table %s: %w", table.Name, err)
		}

		slog.Debug("Generated test", "filename", filename)
	}

	return nil
}

// writeTemplate writes a template to a file
func (g *Generator) writeTemplate(tmpl *template.Template, filepath string, data interface{}) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// getPrimaryKeyType returns the Go type of the primary key
func (g *Generator) getPrimaryKeyType(table introspector.Table) string {
	for _, col := range table.Columns {
		if col.IsPrimaryKey {
			return col.GoType
		}
	}
	return "interface{}"
}

// getPrimaryKeyColumn returns the name of the primary key column
func (g *Generator) getPrimaryKeyColumn(table introspector.Table) string {
	for _, col := range table.Columns {
		if col.IsPrimaryKey {
			return col.Name
		}
	}
	return "id"
}

// Helper functions for naming conventions

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
