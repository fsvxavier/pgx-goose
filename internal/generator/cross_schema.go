package generator

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// CrossSchemaGenerator handles cross-schema relationships and code generation
type CrossSchemaGenerator struct {
	*Generator
	schemas         map[string]*introspector.Schema
	crossReferences map[string][]CrossReference
	schemaMutex     sync.RWMutex
}

// CrossReference represents a reference between schemas
type CrossReference struct {
	SourceSchema   string
	SourceTable    string
	SourceColumn   string
	TargetSchema   string
	TargetTable    string
	TargetColumn   string
	RelationType   RelationType
	ForeignKeyName string
}

// RelationType defines the type of cross-schema relationship
type RelationType int

const (
	OneToOne RelationType = iota
	OneToMany
	ManyToOne
	ManyToMany
)

// SchemaConfig represents configuration for a specific schema
type SchemaConfig struct {
	Name         string   `yaml:"name" json:"name"`
	DSN          string   `yaml:"dsn" json:"dsn"`
	OutputDir    string   `yaml:"output_dir" json:"output_dir"`
	PackageName  string   `yaml:"package_name" json:"package_name"`
	Tables       []string `yaml:"tables" json:"tables"`
	IgnoreTables []string `yaml:"ignore_tables" json:"ignore_tables"`
}

// MultiSchemaConfig represents configuration for multiple schemas
type MultiSchemaConfig struct {
	Schemas            []SchemaConfig `yaml:"schemas" json:"schemas"`
	EnableCrossSchema  bool           `yaml:"enable_cross_schema" json:"enable_cross_schema"`
	CrossSchemaPackage string         `yaml:"cross_schema_package" json:"cross_schema_package"`
}

// NewCrossSchemaGenerator creates a new cross-schema generator
func NewCrossSchemaGenerator(cfg *config.Config) *CrossSchemaGenerator {
	return &CrossSchemaGenerator{
		Generator:       New(cfg),
		schemas:         make(map[string]*introspector.Schema),
		crossReferences: make(map[string][]CrossReference),
	}
}

// GenerateCrossSchema generates code for multiple schemas with cross-references
func (csg *CrossSchemaGenerator) GenerateCrossSchema(multiConfig *MultiSchemaConfig) error {
	slog.Info("Starting cross-schema code generation", "schemas", len(multiConfig.Schemas))

	// Phase 1: Introspect all schemas
	if err := csg.introspectAllSchemas(multiConfig); err != nil {
		return fmt.Errorf("failed to introspect schemas: %w", err)
	}

	// Phase 2: Discover cross-schema relationships
	if multiConfig.EnableCrossSchema {
		if err := csg.discoverCrossReferences(); err != nil {
			return fmt.Errorf("failed to discover cross-references: %w", err)
		}
	}

	// Phase 3: Generate code for each schema
	for _, schemaConfig := range multiConfig.Schemas {
		if err := csg.generateSchemaCode(schemaConfig, multiConfig); err != nil {
			return fmt.Errorf("failed to generate code for schema %s: %w",
				schemaConfig.Name, err)
		}
	}

	// Phase 4: Generate cross-schema utilities if enabled
	if multiConfig.EnableCrossSchema {
		if err := csg.generateCrossSchemaUtils(multiConfig); err != nil {
			return fmt.Errorf("failed to generate cross-schema utilities: %w", err)
		}
	}

	slog.Info("Cross-schema code generation completed successfully")
	return nil
}

// introspectAllSchemas introspects all configured schemas
func (csg *CrossSchemaGenerator) introspectAllSchemas(multiConfig *MultiSchemaConfig) error {
	csg.schemaMutex.Lock()
	defer csg.schemaMutex.Unlock()

	for _, schemaConfig := range multiConfig.Schemas {
		slog.Info("Introspecting schema", "name", schemaConfig.Name)

		// Create introspector for this schema
		inspector := introspector.New(schemaConfig.DSN, schemaConfig.Name)

		// Introspect schema
		var tablesToProcess []string
		if len(schemaConfig.Tables) > 0 {
			tablesToProcess = csg.filterTables(schemaConfig.Tables, schemaConfig.IgnoreTables)
		}

		schema, err := inspector.IntrospectSchema(tablesToProcess)
		if err != nil {
			return fmt.Errorf("failed to introspect schema %s: %w", schemaConfig.Name, err)
		}

		// Filter ignored tables if needed
		if len(schemaConfig.IgnoreTables) > 0 && len(schemaConfig.Tables) == 0 {
			schema.Tables = csg.filterIgnoredTables(schema.Tables, schemaConfig.IgnoreTables)
		}

		csg.schemas[schemaConfig.Name] = schema
		slog.Info("Schema introspected", "name", schemaConfig.Name, "tables", len(schema.Tables))
	}

	return nil
}

// discoverCrossReferences discovers relationships between schemas
func (csg *CrossSchemaGenerator) discoverCrossReferences() error {
	slog.Info("Discovering cross-schema references")

	for schemaName, schema := range csg.schemas {
		for _, table := range schema.Tables {
			for _, fk := range table.ForeignKeys {
				// Check if foreign key references a different schema
				if crossRef := csg.parseCrossSchemaReference(schemaName, table.Name, fk); crossRef != nil {
					csg.addCrossReference(*crossRef)
				}
			}
		}
	}

	totalRefs := 0
	for schema, refs := range csg.crossReferences {
		totalRefs += len(refs)
		slog.Info("Cross-schema references found", "schema", schema, "count", len(refs))
	}

	slog.Info("Cross-schema reference discovery completed", "total_references", totalRefs)
	return nil
}

// parseCrossSchemaReference parses a foreign key to check for cross-schema reference
func (csg *CrossSchemaGenerator) parseCrossSchemaReference(sourceSchema, sourceTable string, fk introspector.ForeignKey) *CrossReference {
	// Parse referenced table for schema.table format
	parts := strings.Split(fk.ReferencedTable, ".")
	if len(parts) != 2 {
		return nil // Not a cross-schema reference
	}

	targetSchema := parts[0]
	targetTable := parts[1]

	// Check if target schema exists in our schemas
	if _, exists := csg.schemas[targetSchema]; !exists {
		return nil
	}

	// Check if target schema is different from source
	if targetSchema == sourceSchema {
		return nil
	}

	return &CrossReference{
		SourceSchema:   sourceSchema,
		SourceTable:    sourceTable,
		SourceColumn:   fk.Column,
		TargetSchema:   targetSchema,
		TargetTable:    targetTable,
		TargetColumn:   fk.ReferencedColumn,
		RelationType:   ManyToOne, // Default, could be enhanced
		ForeignKeyName: fk.Name,
	}
}

// addCrossReference adds a cross-schema reference
func (csg *CrossSchemaGenerator) addCrossReference(ref CrossReference) {
	csg.crossReferences[ref.SourceSchema] = append(csg.crossReferences[ref.SourceSchema], ref)
}

// generateSchemaCode generates code for a specific schema
func (csg *CrossSchemaGenerator) generateSchemaCode(schemaConfig SchemaConfig, multiConfig *MultiSchemaConfig) error {
	slog.Info("Generating code for schema", "name", schemaConfig.Name)

	schema := csg.schemas[schemaConfig.Name]
	if schema == nil {
		return fmt.Errorf("schema %s not found", schemaConfig.Name)
	}

	// Create a modified config for this schema
	cfg := *csg.config
	cfg.Schema = schemaConfig.Name

	if schemaConfig.OutputDir != "" {
		cfg.OutputDir = schemaConfig.OutputDir
		cfg.OutputDirs.Base = schemaConfig.OutputDir
		cfg.ApplyDefaults()
	}

	// Create generator for this schema
	generator := New(&cfg)

	// Create output directories
	if err := generator.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate regular code
	if err := generator.Generate(schema); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Generate cross-schema relationship code
	if multiConfig.EnableCrossSchema {
		if err := csg.generateCrossSchemaRelationships(schemaConfig, multiConfig); err != nil {
			return fmt.Errorf("failed to generate cross-schema relationships: %w", err)
		}
	}

	return nil
}

// generateCrossSchemaRelationships generates code for cross-schema relationships
func (csg *CrossSchemaGenerator) generateCrossSchemaRelationships(schemaConfig SchemaConfig, multiConfig *MultiSchemaConfig) error {
	refs := csg.crossReferences[schemaConfig.Name]
	if len(refs) == 0 {
		return nil
	}

	slog.Info("Generating cross-schema relationships", "schema", schemaConfig.Name, "references", len(refs))

	// Generate cross-schema repository methods
	if err := csg.generateCrossSchemaRepositories(schemaConfig, refs, multiConfig); err != nil {
		return err
	}

	// Generate cross-schema models with references
	if err := csg.generateCrossSchemaModels(schemaConfig, refs, multiConfig); err != nil {
		return err
	}

	return nil
}

// generateCrossSchemaRepositories generates repository methods for cross-schema access
func (csg *CrossSchemaGenerator) generateCrossSchemaRepositories(schemaConfig SchemaConfig, refs []CrossReference, multiConfig *MultiSchemaConfig) error {
	// This would generate repository methods that can join across schemas
	// For example: GetUserWithProfile where User is in one schema and Profile in another

	slog.Debug("Generating cross-schema repository methods", "schema", schemaConfig.Name)

	// Template data would include:
	// - Source schema models
	// - Target schema models
	// - Join conditions
	// - Cross-schema imports

	return nil
}

// generateCrossSchemaModels generates models with cross-schema references
func (csg *CrossSchemaGenerator) generateCrossSchemaModels(schemaConfig SchemaConfig, refs []CrossReference, multiConfig *MultiSchemaConfig) error {
	// This would generate models that include references to other schema objects
	// For example: adding methods or fields that reference cross-schema entities

	slog.Debug("Generating cross-schema model enhancements", "schema", schemaConfig.Name)

	return nil
}

// generateCrossSchemaUtils generates utility code for cross-schema operations
func (csg *CrossSchemaGenerator) generateCrossSchemaUtils(multiConfig *MultiSchemaConfig) error {
	slog.Info("Generating cross-schema utilities")

	// Generate cross-schema transaction manager
	if err := csg.generateTransactionManager(multiConfig); err != nil {
		return err
	}

	// Generate cross-schema query builder
	if err := csg.generateQueryBuilder(multiConfig); err != nil {
		return err
	}

	// Generate cross-schema migration utilities
	if err := csg.generateMigrationUtils(multiConfig); err != nil {
		return err
	}

	return nil
}

// generateTransactionManager generates a transaction manager for cross-schema operations
func (csg *CrossSchemaGenerator) generateTransactionManager(multiConfig *MultiSchemaConfig) error {
	slog.Debug("Generating cross-schema transaction manager")

	// Template would generate:
	// - Multi-connection transaction manager
	// - Cross-schema rollback handling
	// - Distributed transaction support

	return nil
}

// generateQueryBuilder generates a query builder for cross-schema joins
func (csg *CrossSchemaGenerator) generateQueryBuilder(multiConfig *MultiSchemaConfig) error {
	slog.Debug("Generating cross-schema query builder")

	// Template would generate:
	// - Cross-schema join builder
	// - Schema-aware query methods
	// - Dynamic schema switching

	return nil
}

// generateMigrationUtils generates migration utilities for cross-schema changes
func (csg *CrossSchemaGenerator) generateMigrationUtils(multiConfig *MultiSchemaConfig) error {
	slog.Debug("Generating cross-schema migration utilities")

	// Template would generate:
	// - Cross-schema migration runner
	// - Dependency-aware migration ordering
	// - Cross-schema foreign key management

	return nil
}

// Helper methods

// filterTables filters tables to include only specified ones, excluding ignored ones
func (csg *CrossSchemaGenerator) filterTables(tables, ignoreTables []string) []string {
	if len(ignoreTables) == 0 {
		return tables
	}

	ignoreMap := make(map[string]bool)
	for _, table := range ignoreTables {
		ignoreMap[strings.ToLower(table)] = true
	}

	var filtered []string
	for _, table := range tables {
		if !ignoreMap[strings.ToLower(table)] {
			filtered = append(filtered, table)
		}
	}

	return filtered
}

// filterIgnoredTables removes ignored tables from the list
func (csg *CrossSchemaGenerator) filterIgnoredTables(tables []introspector.Table, ignoreTables []string) []introspector.Table {
	if len(ignoreTables) == 0 {
		return tables
	}

	ignoreMap := make(map[string]bool)
	for _, table := range ignoreTables {
		ignoreMap[strings.ToLower(table)] = true
	}

	var filtered []introspector.Table
	for _, table := range tables {
		if !ignoreMap[strings.ToLower(table.Name)] {
			filtered = append(filtered, table)
		}
	}

	return filtered
}

// GetCrossReferences returns cross-references for a specific schema
func (csg *CrossSchemaGenerator) GetCrossReferences(schemaName string) []CrossReference {
	csg.schemaMutex.RLock()
	defer csg.schemaMutex.RUnlock()

	return csg.crossReferences[schemaName]
}

// GetAllSchemas returns all introspected schemas
func (csg *CrossSchemaGenerator) GetAllSchemas() map[string]*introspector.Schema {
	csg.schemaMutex.RLock()
	defer csg.schemaMutex.RUnlock()

	result := make(map[string]*introspector.Schema)
	for name, schema := range csg.schemas {
		result[name] = schema
	}

	return result
}
