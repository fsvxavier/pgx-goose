package generator

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// IncrementalGenerator handles incremental code generation
type IncrementalGenerator struct {
	*Generator
	metadataFile string
	metadata     *GenerationMetadata
}

// GenerationMetadata stores metadata about the last generation
type GenerationMetadata struct {
	LastGeneration time.Time                    `json:"last_generation"`
	SchemaHash     string                       `json:"schema_hash"`
	ConfigHash     string                       `json:"config_hash"`
	TableHashes    map[string]string            `json:"table_hashes"`
	FileHashes     map[string]string            `json:"file_hashes"`
	GeneratedFiles map[string]GeneratedFileInfo `json:"generated_files"`
	Version        string                       `json:"version"`
}

// GeneratedFileInfo contains information about a generated file
type GeneratedFileInfo struct {
	Path           string    `json:"path"`
	Hash           string    `json:"hash"`
	Size           int64     `json:"size"`
	ModTime        time.Time `json:"mod_time"`
	TableName      string    `json:"table_name"`
	GenerationType string    `json:"generation_type"`
}

// ChangeDetector detects changes in database schema and configuration
type ChangeDetector struct {
	metadata *GenerationMetadata
}

// TableChange represents a change to a table
type TableChange struct {
	TableName  string
	ChangeType ChangeType
	OldHash    string
	NewHash    string
}

// ChangeType represents the type of change
type ChangeType int

const (
	TableAdded ChangeType = iota
	TableModified
	TableRemoved
	TableUnchanged
)

// NewIncrementalGenerator creates a new incremental generator
func NewIncrementalGenerator(cfg *config.Config) *IncrementalGenerator {
	metadataFile := filepath.Join(cfg.GetBaseDir(), ".pgx-goose-metadata.json")

	ig := &IncrementalGenerator{
		Generator:    New(cfg),
		metadataFile: metadataFile,
		metadata:     &GenerationMetadata{},
	}

	// Load existing metadata if available
	ig.loadMetadata()

	return ig
}

// GenerateIncremental performs incremental code generation
func (ig *IncrementalGenerator) GenerateIncremental(schema *introspector.Schema) error {
	slog.Info("Starting incremental code generation")

	// Create directories first
	if err := ig.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Detect changes
	changes, err := ig.detectChanges(schema)
	if err != nil {
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	if len(changes) == 0 {
		slog.Info("No changes detected, skipping generation")
		return nil
	}

	slog.Info("Changes detected", "count", len(changes))
	for _, change := range changes {
		slog.Info("Table change detected",
			"table", change.TableName,
			"type", change.ChangeType)
	}

	// Generate only changed tables
	changedTables := ig.getChangedTables(schema, changes)
	if len(changedTables) == 0 {
		slog.Info("No tables need regeneration")
		return nil
	}

	// Create schema with only changed tables
	incrementalSchema := &introspector.Schema{
		Tables: changedTables,
	}

	// Remove obsolete files first
	if err := ig.removeObsoleteFiles(changes); err != nil {
		slog.Warn("Failed to remove obsolete files", "error", err)
	}

	// Generate code for changed tables
	if err := ig.Generator.Generate(incrementalSchema); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Update metadata
	if err := ig.updateMetadata(schema); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	slog.Info("Incremental code generation completed",
		"changed_tables", len(changedTables))

	return nil
}

// detectChanges detects changes between current schema and last generation
func (ig *IncrementalGenerator) detectChanges(schema *introspector.Schema) ([]TableChange, error) {
	var changes []TableChange

	// Calculate current schema hash
	currentSchemaHash, err := ig.calculateSchemaHash(schema)
	if err != nil {
		return nil, err
	}

	// Calculate current config hash
	currentConfigHash, err := ig.calculateConfigHash()
	if err != nil {
		return nil, err
	}

	// Check if this is the first generation or config changed
	if ig.metadata.SchemaHash == "" || ig.metadata.ConfigHash != currentConfigHash {
		slog.Info("First generation or config changed, regenerating all tables")
		for _, table := range schema.Tables {
			changes = append(changes, TableChange{
				TableName:  table.Name,
				ChangeType: TableAdded,
				NewHash:    ig.calculateTableHash(table),
			})
		}
		return changes, nil
	}

	// Check if overall schema changed
	if ig.metadata.SchemaHash != currentSchemaHash {
		// Detailed table comparison
		currentTableHashes := make(map[string]string)
		for _, table := range schema.Tables {
			currentTableHashes[table.Name] = ig.calculateTableHash(table)
		}

		// Find new and modified tables
		for tableName, currentHash := range currentTableHashes {
			if oldHash, exists := ig.metadata.TableHashes[tableName]; exists {
				if oldHash != currentHash {
					changes = append(changes, TableChange{
						TableName:  tableName,
						ChangeType: TableModified,
						OldHash:    oldHash,
						NewHash:    currentHash,
					})
				}
			} else {
				changes = append(changes, TableChange{
					TableName:  tableName,
					ChangeType: TableAdded,
					NewHash:    currentHash,
				})
			}
		}

		// Find removed tables
		for tableName, oldHash := range ig.metadata.TableHashes {
			if _, exists := currentTableHashes[tableName]; !exists {
				changes = append(changes, TableChange{
					TableName:  tableName,
					ChangeType: TableRemoved,
					OldHash:    oldHash,
				})
			}
		}
	}

	return changes, nil
}

// getChangedTables returns tables that need regeneration
func (ig *IncrementalGenerator) getChangedTables(schema *introspector.Schema, changes []TableChange) []introspector.Table {
	changedTableNames := make(map[string]bool)

	for _, change := range changes {
		if change.ChangeType == TableAdded || change.ChangeType == TableModified {
			changedTableNames[change.TableName] = true
		}
	}

	var changedTables []introspector.Table
	for _, table := range schema.Tables {
		if changedTableNames[table.Name] {
			changedTables = append(changedTables, table)
		}
	}

	return changedTables
}

// removeObsoleteFiles removes files for deleted tables
func (ig *IncrementalGenerator) removeObsoleteFiles(changes []TableChange) error {
	for _, change := range changes {
		if change.ChangeType == TableRemoved {
			if err := ig.removeTableFiles(change.TableName); err != nil {
				slog.Warn("Failed to remove files for deleted table",
					"table", change.TableName, "error", err)
			}
		}
	}
	return nil
}

// removeTableFiles removes all generated files for a specific table
func (ig *IncrementalGenerator) removeTableFiles(tableName string) error {
	// Get all generated files for this table from metadata
	for filePath, fileInfo := range ig.metadata.GeneratedFiles {
		if fileInfo.TableName == tableName {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove file %s: %w", filePath, err)
			}
			slog.Debug("Removed obsolete file", "path", filePath, "table", tableName)
		}
	}
	return nil
}

// calculateSchemaHash calculates a hash for the entire schema
func (ig *IncrementalGenerator) calculateSchemaHash(schema *introspector.Schema) (string, error) {
	hasher := sha256.New()

	// Sort tables by name for consistent hashing
	tableHashes := make([]string, 0, len(schema.Tables))
	for _, table := range schema.Tables {
		tableHashes = append(tableHashes, ig.calculateTableHash(table))
	}

	for _, hash := range tableHashes {
		hasher.Write([]byte(hash))
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// calculateTableHash calculates a hash for a single table
func (ig *IncrementalGenerator) calculateTableHash(table introspector.Table) string {
	hasher := sha256.New()

	// Hash table name
	hasher.Write([]byte(table.Name))

	// Hash columns
	for _, col := range table.Columns {
		hasher.Write([]byte(fmt.Sprintf("%s:%s:%t:%t",
			col.Name, col.Type, col.IsNullable, col.IsPrimaryKey)))
	}

	// Hash foreign keys
	for _, fk := range table.ForeignKeys {
		hasher.Write([]byte(fmt.Sprintf("%s:%s:%s:%s",
			fk.Column, fk.ReferencedTable,
			fk.ReferencedColumn, fk.Name)))
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// calculateConfigHash calculates a hash for the configuration
func (ig *IncrementalGenerator) calculateConfigHash() (string, error) {
	hasher := sha256.New()

	// Hash relevant config fields that affect generation
	configData := fmt.Sprintf("%s:%s:%t:%t:%s",
		ig.config.TemplateDir,
		ig.config.MockProvider,
		ig.config.WithTests,
		ig.config.OutputDir != "",
		fmt.Sprintf("%v", ig.config.Tables))

	hasher.Write([]byte(configData))
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// loadMetadata loads generation metadata from file
func (ig *IncrementalGenerator) loadMetadata() error {
	if _, err := os.Stat(ig.metadataFile); os.IsNotExist(err) {
		// First run, initialize empty metadata
		ig.metadata = &GenerationMetadata{
			TableHashes:    make(map[string]string),
			FileHashes:     make(map[string]string),
			GeneratedFiles: make(map[string]GeneratedFileInfo),
			Version:        "1.0",
		}
		return nil
	}

	data, err := os.ReadFile(ig.metadataFile)
	if err != nil {
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	if err := json.Unmarshal(data, ig.metadata); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	slog.Debug("Loaded generation metadata",
		"last_generation", ig.metadata.LastGeneration,
		"tables", len(ig.metadata.TableHashes))

	return nil
}

// updateMetadata updates and saves generation metadata
func (ig *IncrementalGenerator) updateMetadata(schema *introspector.Schema) error {
	// Update metadata
	ig.metadata.LastGeneration = time.Now()

	schemaHash, err := ig.calculateSchemaHash(schema)
	if err != nil {
		return err
	}
	ig.metadata.SchemaHash = schemaHash

	configHash, err := ig.calculateConfigHash()
	if err != nil {
		return err
	}
	ig.metadata.ConfigHash = configHash

	// Update table hashes
	ig.metadata.TableHashes = make(map[string]string)
	for _, table := range schema.Tables {
		ig.metadata.TableHashes[table.Name] = ig.calculateTableHash(table)
	}

	// Update file information
	if err := ig.updateFileMetadata(); err != nil {
		return err
	}

	// Save metadata
	return ig.saveMetadata()
}

// updateFileMetadata updates metadata for generated files
func (ig *IncrementalGenerator) updateFileMetadata() error {
	// This would scan generated directories and update file information
	// Implementation depends on how files are organized

	outputDirs := ig.config.GetAllOutputDirs()
	for _, dir := range outputDirs {
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && (filepath.Ext(path) == ".go" || filepath.Ext(path) == ".sql") {
				// Calculate file hash
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				hasher := sha256.New()
				hasher.Write(data)
				hash := fmt.Sprintf("%x", hasher.Sum(nil))

				// Store file metadata
				ig.metadata.FileHashes[path] = hash
				ig.metadata.GeneratedFiles[path] = GeneratedFileInfo{
					Path:    path,
					Hash:    hash,
					Size:    info.Size(),
					ModTime: info.ModTime(),
					// TableName and GenerationType would be determined from file path/content
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// saveMetadata saves generation metadata to file
func (ig *IncrementalGenerator) saveMetadata() error {
	data, err := json.MarshalIndent(ig.metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(ig.metadataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	slog.Debug("Saved generation metadata", "file", ig.metadataFile)
	return nil
}

// ForceRegeneration forces regeneration of all files by clearing metadata
func (ig *IncrementalGenerator) ForceRegeneration() error {
	ig.metadata = &GenerationMetadata{
		TableHashes:    make(map[string]string),
		FileHashes:     make(map[string]string),
		GeneratedFiles: make(map[string]GeneratedFileInfo),
		Version:        "1.0",
	}

	// Remove metadata file
	if err := os.Remove(ig.metadataFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	slog.Info("Forced regeneration - metadata cleared")
	return nil
}
