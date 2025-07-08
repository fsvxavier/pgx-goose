package introspector

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ServiceConfig contains configuration for the introspector service
type ServiceConfig struct {
	Pool   *pgxpool.Pool
	Schema string
	Logger *slog.Logger
}

// IntrospectorService implements enhanced introspection with observability
type IntrospectorService struct {
	pool   *pgxpool.Pool
	schema string
	logger *slog.Logger
}

// NewIntrospectorService creates a new introspector service with dependency injection
func NewIntrospectorService(config ServiceConfig) *IntrospectorService {
	if config.Schema == "" {
		config.Schema = "public"
	}

	return &IntrospectorService{
		pool:   config.Pool,
		schema: config.Schema,
		logger: config.Logger,
	}
}

func (i *IntrospectorService) IntrospectSchema(ctx context.Context, tables []string) (*Schema, error) {
	i.logger.Info("Starting schema introspection",
		"schema", i.schema,
		"table_count", len(tables))

	// Test connection
	if err := i.pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	schema := &Schema{}

	// Get all tables if none specified
	if len(tables) == 0 {
		var err error
		tables, err = i.GetAllTables(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tables: %w", err)
		}
	}

	// Process each table
	for _, tableName := range tables {
		table, err := i.introspectTable(ctx, tableName)
		if err != nil {
			i.logger.Error("Failed to introspect table",
				"table", tableName,
				"error", err)
			continue
		}
		schema.Tables = append(schema.Tables, *table)

		i.logger.Debug("Table introspected",
			"table", tableName,
			"columns", len(table.Columns),
			"indexes", len(table.Indexes),
			"foreign_keys", len(table.ForeignKeys))
	}

	i.logger.Info("Schema introspection completed",
		"schema", i.schema,
		"tables_processed", len(schema.Tables))

	return schema, nil
}

func (i *IntrospectorService) GetAllTables(ctx context.Context) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name`

	rows, err := i.pool.Query(ctx, query, i.schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	i.logger.Debug("Retrieved tables",
		"schema", i.schema,
		"count", len(tables))

	return tables, nil
}

func (i *IntrospectorService) introspectTable(ctx context.Context, tableName string) (*Table, error) {
	table := &Table{Name: tableName}

	// Get table comment
	if err := i.getTableComment(ctx, table); err != nil {
		i.logger.Warn("Failed to get table comment",
			"table", tableName,
			"error", err)
	}

	// Get columns
	if err := i.getTableColumns(ctx, table); err != nil {
		return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}

	// Get primary keys
	if err := i.getTablePrimaryKeys(ctx, table); err != nil {
		return nil, fmt.Errorf("failed to get primary keys for table %s: %w", tableName, err)
	}

	// Get indexes
	if err := i.getTableIndexes(ctx, table); err != nil {
		i.logger.Warn("Failed to get indexes",
			"table", tableName,
			"error", err)
	}

	// Get foreign keys
	if err := i.getTableForeignKeys(ctx, table); err != nil {
		i.logger.Warn("Failed to get foreign keys",
			"table", tableName,
			"error", err)
	}

	return table, nil
}

func (i *IntrospectorService) getTableComment(ctx context.Context, table *Table) error {
	query := `
		SELECT obj_description(c.oid) 
		FROM pg_class c 
		JOIN pg_namespace n ON n.oid = c.relnamespace 
		WHERE c.relname = $1 AND n.nspname = $2`

	row := i.pool.QueryRow(ctx, query, table.Name, i.schema)
	var comment *string
	if err := row.Scan(&comment); err != nil {
		return err
	}

	if comment != nil {
		table.Comment = *comment
	}

	return nil
}

func (i *IntrospectorService) getTableColumns(ctx context.Context, table *Table) error {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.ordinal_position,
			COALESCE(pgd.description, '') as comment
		FROM information_schema.columns c
		LEFT JOIN pg_class pgc ON pgc.relname = c.table_name
		LEFT JOIN pg_namespace pgn ON pgn.oid = pgc.relnamespace AND pgn.nspname = c.table_schema
		LEFT JOIN pg_attribute pga ON pga.attrelid = pgc.oid AND pga.attname = c.column_name
		LEFT JOIN pg_description pgd ON pgd.objoid = pgc.oid AND pgd.objsubid = pga.attnum
		WHERE c.table_name = $1 AND c.table_schema = $2
		ORDER BY c.ordinal_position`

	rows, err := i.pool.Query(ctx, query, table.Name, i.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var column Column
		var dataType string
		var isNullable string
		var position int

		err := rows.Scan(
			&column.Name,
			&dataType,
			&isNullable,
			&column.DefaultValue,
			&position,
			&column.Comment,
		)
		if err != nil {
			return err
		}

		column.Type = dataType
		column.GoType = mapPostgresToGoType(dataType, isNullable == "YES")
		column.IsNullable = isNullable == "YES"
		column.Position = position

		table.Columns = append(table.Columns, column)
	}

	return rows.Err()
}

func (i *IntrospectorService) getTablePrimaryKeys(ctx context.Context, table *Table) error {
	query := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass AND i.indisprimary
		ORDER BY a.attnum`

	tableName := fmt.Sprintf("%s.%s", i.schema, table.Name)
	rows, err := i.pool.Query(ctx, query, tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pkColumn string
		if err := rows.Scan(&pkColumn); err != nil {
			return err
		}
		table.PrimaryKeys = append(table.PrimaryKeys, pkColumn)

		// Mark column as primary key
		for j := range table.Columns {
			if table.Columns[j].Name == pkColumn {
				table.Columns[j].IsPrimaryKey = true
				break
			}
		}
	}

	return rows.Err()
}

func (i *IntrospectorService) getTableIndexes(ctx context.Context, table *Table) error {
	query := `
		SELECT 
			i.relname as index_name,
			array_agg(a.attname ORDER BY a.attnum) as columns,
			ix.indisunique
		FROM pg_class t
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relname = $1 AND n.nspname = $2 AND NOT ix.indisprimary
		GROUP BY i.relname, ix.indisunique
		ORDER BY i.relname`

	rows, err := i.pool.Query(ctx, query, table.Name, i.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var index Index
		var columns []string

		err := rows.Scan(&index.Name, &columns, &index.IsUnique)
		if err != nil {
			return err
		}

		index.Columns = columns
		table.Indexes = append(table.Indexes, index)
	}

	return rows.Err()
}

func (i *IntrospectorService) getTableForeignKeys(ctx context.Context, table *Table) error {
	query := `
		SELECT 
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_name = $1
			AND tc.table_schema = $2`

	rows, err := i.pool.Query(ctx, query, table.Name, i.schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var fk ForeignKey
		err := rows.Scan(
			&fk.Name,
			&fk.Column,
			&fk.ReferencedTable,
			&fk.ReferencedColumn,
		)
		if err != nil {
			return err
		}
		table.ForeignKeys = append(table.ForeignKeys, fk)
	}

	return rows.Err()
}

func (i *IntrospectorService) Close() error {
	if i.pool != nil {
		i.pool.Close()
	}
	return nil
}

// mapPostgresToGoType maps PostgreSQL data types to Go types
func mapPostgresToGoType(postgresType string, isNullable bool) string {
	var goType string

	switch postgresType {
	case "integer", "int4":
		goType = "int32"
	case "bigint", "int8":
		goType = "int64"
	case "smallint", "int2":
		goType = "int16"
	case "text", "varchar", "character varying":
		goType = "string"
	case "boolean", "bool":
		goType = "bool"
	case "timestamp", "timestamp without time zone":
		goType = "time.Time"
	case "timestamp with time zone", "timestamptz":
		goType = "time.Time"
	case "date":
		goType = "time.Time"
	case "uuid":
		goType = "string"
	case "json", "jsonb":
		goType = "json.RawMessage"
	case "decimal", "numeric":
		goType = "float64"
	case "real", "float4":
		goType = "float32"
	case "double precision", "float8":
		goType = "float64"
	case "bytea":
		goType = "[]byte"
	default:
		goType = "interface{}"
	}

	if isNullable {
		switch goType {
		case "string":
			return "*string"
		case "int32":
			return "*int32"
		case "int64":
			return "*int64"
		case "int16":
			return "*int16"
		case "bool":
			return "*bool"
		case "float32":
			return "*float32"
		case "float64":
			return "*float64"
		case "time.Time":
			return "*time.Time"
		case "[]byte":
			return "[]byte"
		default:
			return goType
		}
	}

	return goType
}
