package generator

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// MigrationGenerator handles database migration generation
type MigrationGenerator struct {
	config       *config.Config
	optimizer    *TemplateOptimizer
	migrationDir string
}

// Migration represents a database migration
type Migration struct {
	Version      string
	Name         string
	UpSQL        string
	DownSQL      string
	Description  string
	Timestamp    time.Time
	Dependencies []string
}

// SchemaDiff represents differences between two schema versions
type SchemaDiff struct {
	AddedTables        []introspector.Table
	DroppedTables      []string
	ModifiedTables     []TableDiff
	AddedColumns       map[string][]introspector.Column
	DroppedColumns     map[string][]string
	ModifiedColumns    map[string][]ColumnDiff
	AddedIndexes       map[string][]introspector.Index
	DroppedIndexes     map[string][]string
	AddedForeignKeys   map[string][]introspector.ForeignKey
	DroppedForeignKeys map[string][]string
}

// TableDiff represents changes to a table
type TableDiff struct {
	TableName  string
	OldComment string
	NewComment string
	Changes    []TableChangeItem
}

// TableChangeItem represents a specific change to a table
type TableChangeItem struct {
	Type string
	Old  string
	New  string
}

// ColumnDiff represents changes to a column
type ColumnDiff struct {
	ColumnName  string
	OldType     string
	NewType     string
	OldNullable bool
	NewNullable bool
	OldDefault  *string
	NewDefault  *string
	ChangeType  ColumnChangeType
}

// ColumnChangeType represents the type of column change
type ColumnChangeType int

const (
	ColumnTypeChanged ColumnChangeType = iota
	ColumnNullabilityChanged
	ColumnDefaultChanged
	ColumnRenamed
)

// MigrationConfig represents migration generation configuration
type MigrationConfig struct {
	MigrationDir    string `yaml:"migration_dir" json:"migration_dir"`
	MigrationFormat string `yaml:"migration_format" json:"migration_format"` // "goose", "migrate", "custom"
	AutoGenerate    bool   `yaml:"auto_generate" json:"auto_generate"`
	IncludeDrops    bool   `yaml:"include_drops" json:"include_drops"`
	IncludeData     bool   `yaml:"include_data" json:"include_data"`
	BatchSize       int    `yaml:"batch_size" json:"batch_size"`
	SafeMode        bool   `yaml:"safe_mode" json:"safe_mode"`
}

// NewMigrationGenerator creates a new migration generator
func NewMigrationGenerator(cfg *config.Config) *MigrationGenerator {
	migrationDir := filepath.Join(cfg.GetBaseDir(), "migrations")

	return &MigrationGenerator{
		config:       cfg,
		optimizer:    NewTemplateOptimizer(50),
		migrationDir: migrationDir,
	}
}

// GenerateMigrations generates migrations based on schema differences
func (mg *MigrationGenerator) GenerateMigrations(oldSchema, newSchema *introspector.Schema, migrationConfig *MigrationConfig) error {
	slog.Info("Starting migration generation")

	// Ensure migration directory exists
	if err := os.MkdirAll(mg.migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	// Calculate schema differences
	diff, err := mg.calculateSchemaDiff(oldSchema, newSchema)
	if err != nil {
		return fmt.Errorf("failed to calculate schema diff: %w", err)
	}

	// Check if there are any changes
	if mg.isDiffEmpty(diff) {
		slog.Info("No schema changes detected, no migrations generated")
		return nil
	}

	// Generate migrations
	migrations, err := mg.generateMigrationsFromDiff(diff, migrationConfig)
	if err != nil {
		return fmt.Errorf("failed to generate migrations: %w", err)
	}

	// Write migrations to files
	for _, migration := range migrations {
		if err := mg.writeMigrationFiles(migration, migrationConfig); err != nil {
			return fmt.Errorf("failed to write migration %s: %w", migration.Name, err)
		}
	}

	slog.Info("Migration generation completed", "migrations_created", len(migrations))
	return nil
}

// calculateSchemaDiff calculates differences between two schemas
func (mg *MigrationGenerator) calculateSchemaDiff(oldSchema, newSchema *introspector.Schema) (*SchemaDiff, error) {
	diff := &SchemaDiff{
		AddedColumns:       make(map[string][]introspector.Column),
		DroppedColumns:     make(map[string][]string),
		ModifiedColumns:    make(map[string][]ColumnDiff),
		AddedIndexes:       make(map[string][]introspector.Index),
		DroppedIndexes:     make(map[string][]string),
		AddedForeignKeys:   make(map[string][]introspector.ForeignKey),
		DroppedForeignKeys: make(map[string][]string),
	}

	// Create lookup maps for old schema
	oldTables := make(map[string]introspector.Table)
	if oldSchema != nil {
		for _, table := range oldSchema.Tables {
			oldTables[table.Name] = table
		}
	}

	// Create lookup maps for new schema
	newTables := make(map[string]introspector.Table)
	for _, table := range newSchema.Tables {
		newTables[table.Name] = table
	}

	// Find added and modified tables
	for tableName, newTable := range newTables {
		if oldTable, exists := oldTables[tableName]; exists {
			// Table exists in both - check for modifications
			if tableDiff := mg.compareTable(oldTable, newTable); tableDiff != nil {
				diff.ModifiedTables = append(diff.ModifiedTables, *tableDiff)
			}

			// Compare columns
			mg.compareColumns(tableName, oldTable, newTable, diff)

			// Compare indexes
			mg.compareIndexes(tableName, oldTable, newTable, diff)

			// Compare foreign keys
			mg.compareForeignKeys(tableName, oldTable, newTable, diff)
		} else {
			// New table
			diff.AddedTables = append(diff.AddedTables, newTable)
		}
	}

	// Find dropped tables
	for tableName := range oldTables {
		if _, exists := newTables[tableName]; !exists {
			diff.DroppedTables = append(diff.DroppedTables, tableName)
		}
	}

	return diff, nil
}

// compareTable compares two tables for differences
func (mg *MigrationGenerator) compareTable(oldTable, newTable introspector.Table) *TableDiff {
	var changes []TableChangeItem

	if oldTable.Comment != newTable.Comment {
		changes = append(changes, TableChangeItem{
			Type: "comment_changed",
			Old:  oldTable.Comment,
			New:  newTable.Comment,
		})
	}

	if len(changes) == 0 {
		return nil
	}

	return &TableDiff{
		TableName:  newTable.Name,
		OldComment: oldTable.Comment,
		NewComment: newTable.Comment,
		Changes:    changes,
	}
}

// compareColumns compares columns between two tables
func (mg *MigrationGenerator) compareColumns(tableName string, oldTable, newTable introspector.Table, diff *SchemaDiff) {
	// Create lookup maps
	oldColumns := make(map[string]introspector.Column)
	for _, col := range oldTable.Columns {
		oldColumns[col.Name] = col
	}

	newColumns := make(map[string]introspector.Column)
	for _, col := range newTable.Columns {
		newColumns[col.Name] = col
	}

	// Find added and modified columns
	for colName, newCol := range newColumns {
		if oldCol, exists := oldColumns[colName]; exists {
			// Column exists - check for modifications
			if colDiff := mg.compareColumn(oldCol, newCol); colDiff != nil {
				diff.ModifiedColumns[tableName] = append(diff.ModifiedColumns[tableName], *colDiff)
			}
		} else {
			// New column
			diff.AddedColumns[tableName] = append(diff.AddedColumns[tableName], newCol)
		}
	}

	// Find dropped columns
	for colName := range oldColumns {
		if _, exists := newColumns[colName]; !exists {
			diff.DroppedColumns[tableName] = append(diff.DroppedColumns[tableName], colName)
		}
	}
}

// compareColumn compares two columns for differences
func (mg *MigrationGenerator) compareColumn(oldCol, newCol introspector.Column) *ColumnDiff {
	var changeType ColumnChangeType
	hasChanges := false

	diff := &ColumnDiff{
		ColumnName:  newCol.Name,
		OldType:     oldCol.Type,
		NewType:     newCol.Type,
		OldNullable: oldCol.IsNullable,
		NewNullable: newCol.IsNullable,
		OldDefault:  oldCol.DefaultValue,
		NewDefault:  newCol.DefaultValue,
	}

	if oldCol.Type != newCol.Type {
		changeType = ColumnTypeChanged
		hasChanges = true
	} else if oldCol.IsNullable != newCol.IsNullable {
		changeType = ColumnNullabilityChanged
		hasChanges = true
	} else if !mg.equalStringPointers(oldCol.DefaultValue, newCol.DefaultValue) {
		changeType = ColumnDefaultChanged
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	diff.ChangeType = changeType
	return diff
}

// compareIndexes compares indexes between two tables
func (mg *MigrationGenerator) compareIndexes(tableName string, oldTable, newTable introspector.Table, diff *SchemaDiff) {
	// Create lookup maps
	oldIndexes := make(map[string]introspector.Index)
	for _, idx := range oldTable.Indexes {
		oldIndexes[idx.Name] = idx
	}

	newIndexes := make(map[string]introspector.Index)
	for _, idx := range newTable.Indexes {
		newIndexes[idx.Name] = idx
	}

	// Find added indexes
	for idxName, newIdx := range newIndexes {
		if _, exists := oldIndexes[idxName]; !exists {
			diff.AddedIndexes[tableName] = append(diff.AddedIndexes[tableName], newIdx)
		}
	}

	// Find dropped indexes
	for idxName := range oldIndexes {
		if _, exists := newIndexes[idxName]; !exists {
			diff.DroppedIndexes[tableName] = append(diff.DroppedIndexes[tableName], idxName)
		}
	}
}

// compareForeignKeys compares foreign keys between two tables
func (mg *MigrationGenerator) compareForeignKeys(tableName string, oldTable, newTable introspector.Table, diff *SchemaDiff) {
	// Create lookup maps
	oldFKs := make(map[string]introspector.ForeignKey)
	for _, fk := range oldTable.ForeignKeys {
		oldFKs[fk.Name] = fk
	}

	newFKs := make(map[string]introspector.ForeignKey)
	for _, fk := range newTable.ForeignKeys {
		newFKs[fk.Name] = fk
	}

	// Find added foreign keys
	for fkName, newFK := range newFKs {
		if _, exists := oldFKs[fkName]; !exists {
			diff.AddedForeignKeys[tableName] = append(diff.AddedForeignKeys[tableName], newFK)
		}
	}

	// Find dropped foreign keys
	for fkName := range oldFKs {
		if _, exists := newFKs[fkName]; !exists {
			diff.DroppedForeignKeys[tableName] = append(diff.DroppedForeignKeys[tableName], fkName)
		}
	}
}

// generateMigrationsFromDiff generates migrations from schema differences
func (mg *MigrationGenerator) generateMigrationsFromDiff(diff *SchemaDiff, config *MigrationConfig) ([]Migration, error) {
	var migrations []Migration
	timestamp := time.Now()

	// Generate table creation migrations
	if len(diff.AddedTables) > 0 {
		migration, err := mg.generateCreateTableMigration(diff.AddedTables, timestamp, config)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
		timestamp = timestamp.Add(time.Second)
	}

	// Generate column addition migrations
	if len(diff.AddedColumns) > 0 {
		migration, err := mg.generateAddColumnMigration(diff.AddedColumns, timestamp, config)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
		timestamp = timestamp.Add(time.Second)
	}

	// Generate column modification migrations
	if len(diff.ModifiedColumns) > 0 {
		migration, err := mg.generateModifyColumnMigration(diff.ModifiedColumns, timestamp, config)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
		timestamp = timestamp.Add(time.Second)
	}

	// Generate index creation migrations
	if len(diff.AddedIndexes) > 0 {
		migration, err := mg.generateCreateIndexMigration(diff.AddedIndexes, timestamp, config)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
		timestamp = timestamp.Add(time.Second)
	}

	// Generate foreign key creation migrations
	if len(diff.AddedForeignKeys) > 0 {
		migration, err := mg.generateCreateForeignKeyMigration(diff.AddedForeignKeys, timestamp, config)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
		timestamp = timestamp.Add(time.Second)
	}

	// Generate drop migrations if enabled
	if config.IncludeDrops {
		// Drop foreign keys first
		if len(diff.DroppedForeignKeys) > 0 {
			migration, err := mg.generateDropForeignKeyMigration(diff.DroppedForeignKeys, timestamp, config)
			if err != nil {
				return nil, err
			}
			migrations = append(migrations, migration)
			timestamp = timestamp.Add(time.Second)
		}

		// Drop indexes
		if len(diff.DroppedIndexes) > 0 {
			migration, err := mg.generateDropIndexMigration(diff.DroppedIndexes, timestamp, config)
			if err != nil {
				return nil, err
			}
			migrations = append(migrations, migration)
			timestamp = timestamp.Add(time.Second)
		}

		// Drop columns
		if len(diff.DroppedColumns) > 0 {
			migration, err := mg.generateDropColumnMigration(diff.DroppedColumns, timestamp, config)
			if err != nil {
				return nil, err
			}
			migrations = append(migrations, migration)
			timestamp = timestamp.Add(time.Second)
		}

		// Drop tables last
		if len(diff.DroppedTables) > 0 {
			migration, err := mg.generateDropTableMigration(diff.DroppedTables, timestamp, config)
			if err != nil {
				return nil, err
			}
			migrations = append(migrations, migration)
		}
	}

	return migrations, nil
}

// generateCreateTableMigration generates a migration for creating tables
func (mg *MigrationGenerator) generateCreateTableMigration(tables []introspector.Table, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	version := timestamp.Format("20060102150405")
	name := fmt.Sprintf("%s_create_tables", version)

	// Generate up SQL
	upSQL, err := mg.generateCreateTableSQL(tables)
	if err != nil {
		return Migration{}, err
	}

	// Generate down SQL
	downSQL := mg.generateDropTableSQL(tables)

	return Migration{
		Version:     version,
		Name:        name,
		UpSQL:       upSQL,
		DownSQL:     downSQL,
		Description: fmt.Sprintf("Create %d tables", len(tables)),
		Timestamp:   timestamp,
	}, nil
}

// Additional migration generation methods would follow similar patterns...

// Helper methods

// isDiffEmpty checks if a schema diff contains no changes
func (mg *MigrationGenerator) isDiffEmpty(diff *SchemaDiff) bool {
	return len(diff.AddedTables) == 0 &&
		len(diff.DroppedTables) == 0 &&
		len(diff.ModifiedTables) == 0 &&
		len(diff.AddedColumns) == 0 &&
		len(diff.DroppedColumns) == 0 &&
		len(diff.ModifiedColumns) == 0 &&
		len(diff.AddedIndexes) == 0 &&
		len(diff.DroppedIndexes) == 0 &&
		len(diff.AddedForeignKeys) == 0 &&
		len(diff.DroppedForeignKeys) == 0
}

// equalStringPointers compares two string pointers for equality
func (mg *MigrationGenerator) equalStringPointers(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// writeMigrationFiles writes migration files to disk
func (mg *MigrationGenerator) writeMigrationFiles(migration Migration, config *MigrationConfig) error {
	switch config.MigrationFormat {
	case "goose":
		return mg.writeGooseMigration(migration)
	case "migrate":
		return mg.writeMigrateMigration(migration)
	default:
		return mg.writeCustomMigration(migration)
	}
}

// writeGooseMigration writes a migration in Goose format
func (mg *MigrationGenerator) writeGooseMigration(migration Migration) error {
	filename := fmt.Sprintf("%s_%s.sql", migration.Version, strings.ReplaceAll(migration.Name, " ", "_"))
	filepath := filepath.Join(mg.migrationDir, filename)

	content := fmt.Sprintf(`-- +goose Up
-- +goose StatementBegin
%s
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
%s
-- +goose StatementEnd
`, migration.UpSQL, migration.DownSQL)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// writeMigrateMigration writes a migration in golang-migrate format
func (mg *MigrationGenerator) writeMigrateMigration(migration Migration) error {
	nameSlug := strings.ReplaceAll(migration.Name, " ", "_")

	// Write up migration
	upFilename := fmt.Sprintf("%s_%s.up.sql", migration.Version, nameSlug)
	upFilepath := filepath.Join(mg.migrationDir, upFilename)
	if err := os.WriteFile(upFilepath, []byte(migration.UpSQL), 0644); err != nil {
		return err
	}

	// Write down migration
	downFilename := fmt.Sprintf("%s_%s.down.sql", migration.Version, nameSlug)
	downFilepath := filepath.Join(mg.migrationDir, downFilename)
	return os.WriteFile(downFilepath, []byte(migration.DownSQL), 0644)
}

// writeCustomMigration writes a migration in custom format
func (mg *MigrationGenerator) writeCustomMigration(migration Migration) error {
	// Implement custom migration format
	return mg.writeGooseMigration(migration) // Default to Goose format
}

// generateCreateTableSQL generates SQL for creating tables
func (mg *MigrationGenerator) generateCreateTableSQL(tables []introspector.Table) (string, error) {
	var sqlParts []string

	for _, table := range tables {
		sql, err := mg.generateSingleCreateTableSQL(table)
		if err != nil {
			return "", err
		}
		sqlParts = append(sqlParts, sql)
	}

	return strings.Join(sqlParts, "\n\n"), nil
}

// generateSingleCreateTableSQL generates SQL for creating a single table
func (mg *MigrationGenerator) generateSingleCreateTableSQL(table introspector.Table) (string, error) {
	tmplContent := `CREATE TABLE {{ .Name }} (
{{- range $i, $col := .Columns }}
{{- if $i }},{{ end }}
    {{ $col.Name }} {{ $col.Type }}{{ if not $col.IsNullable }} NOT NULL{{ end }}{{ if $col.DefaultValue }} DEFAULT {{ $col.DefaultValue }}{{ end }}
{{- end }}
{{- if .PrimaryKeys }},
    PRIMARY KEY ({{ join .PrimaryKeys ", " }})
{{- end }}
);`

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("create_table").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, table); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateDropTableSQL generates SQL for dropping tables
func (mg *MigrationGenerator) generateDropTableSQL(tables []introspector.Table) string {
	var sqlParts []string

	// Drop in reverse order
	for i := len(tables) - 1; i >= 0; i-- {
		sqlParts = append(sqlParts, fmt.Sprintf("DROP TABLE IF EXISTS %s;", tables[i].Name))
	}

	return strings.Join(sqlParts, "\n")
}

// Additional SQL generation methods would be implemented here...

// Placeholder implementations for missing migration types
func (mg *MigrationGenerator) generateAddColumnMigration(columns map[string][]introspector.Column, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for adding columns
	return Migration{}, nil
}

func (mg *MigrationGenerator) generateModifyColumnMigration(columns map[string][]ColumnDiff, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for modifying columns
	return Migration{}, nil
}

func (mg *MigrationGenerator) generateCreateIndexMigration(indexes map[string][]introspector.Index, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for creating indexes
	return Migration{}, nil
}

func (mg *MigrationGenerator) generateCreateForeignKeyMigration(fks map[string][]introspector.ForeignKey, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for creating foreign keys
	return Migration{}, nil
}

func (mg *MigrationGenerator) generateDropForeignKeyMigration(fks map[string][]string, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for dropping foreign keys
	return Migration{}, nil
}

func (mg *MigrationGenerator) generateDropIndexMigration(indexes map[string][]string, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for dropping indexes
	return Migration{}, nil
}

func (mg *MigrationGenerator) generateDropColumnMigration(columns map[string][]string, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for dropping columns
	return Migration{}, nil
}

func (mg *MigrationGenerator) generateDropTableMigration(tables []string, timestamp time.Time, config *MigrationConfig) (Migration, error) {
	// Implementation for dropping tables
	return Migration{}, nil
}
