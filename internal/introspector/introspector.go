package introspector

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Column represents a database column
type Column struct {
	Name         string
	Type         string
	GoType       string
	IsPrimaryKey bool
	IsNullable   bool
	DefaultValue *string
	Comment      string
	Position     int
}

// Index represents a database index
type Index struct {
	Name     string
	Columns  []string
	IsUnique bool
}

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	Name             string
	Column           string
	ReferencedTable  string
	ReferencedColumn string
}

// Table represents a database table
type Table struct {
	Name        string
	Comment     string
	Columns     []Column
	PrimaryKeys []string
	Indexes     []Index
	ForeignKeys []ForeignKey
}

// Schema represents the database schema
type Schema struct {
	Tables []Table
}

// Introspector handles database schema introspection
type Introspector struct {
	dsn    string
	schema string
}

// New creates a new Introspector
func New(dsn, schema string) *Introspector {
	if schema == "" {
		schema = "public"
	}
	return &Introspector{
		dsn:    dsn,
		schema: schema,
	}
}

// IntrospectSchema introspects the database schema
func (i *Introspector) IntrospectSchema(tables []string) (*Schema, error) {
	ctx := context.Background()

	// Connect to database
	pool, err := pgxpool.New(ctx, i.dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	schema := &Schema{}

	// Get all tables if none specified
	if len(tables) == 0 {
		tables, err = i.getAllTables(ctx, pool)
		if err != nil {
			return nil, fmt.Errorf("failed to get tables: %w", err)
		}
	}

	// Process each table
	for _, tableName := range tables {
		table, err := i.introspectTable(ctx, pool, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to introspect table %s: %w", tableName, err)
		}
		schema.Tables = append(schema.Tables, *table)
	}

	return schema, nil
}

// getAllTables returns all table names in the specified schema
func (i *Introspector) getAllTables(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := pool.Query(ctx, query, i.schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// introspectTable introspects a single table
func (i *Introspector) introspectTable(ctx context.Context, pool *pgxpool.Pool, tableName string) (*Table, error) {
	table := &Table{Name: tableName}

	// Get table comment
	comment, err := i.getTableComment(ctx, pool, tableName)
	if err != nil {
		return nil, err
	}
	table.Comment = comment

	// Get columns
	columns, err := i.getColumns(ctx, pool, tableName)
	if err != nil {
		return nil, err
	}
	table.Columns = columns

	// Get primary keys
	primaryKeys, err := i.getPrimaryKeys(ctx, pool, tableName)
	if err != nil {
		return nil, err
	}
	table.PrimaryKeys = primaryKeys

	// Mark primary key columns
	for i := range table.Columns {
		for _, pk := range primaryKeys {
			if table.Columns[i].Name == pk {
				table.Columns[i].IsPrimaryKey = true
				break
			}
		}
	}

	// Get indexes
	indexes, err := i.getIndexes(ctx, pool, tableName)
	if err != nil {
		return nil, err
	}
	table.Indexes = indexes

	// Get foreign keys
	foreignKeys, err := i.getForeignKeys(ctx, pool, tableName)
	if err != nil {
		return nil, err
	}
	table.ForeignKeys = foreignKeys

	return table, nil
}

// getTableComment gets table comment
func (i *Introspector) getTableComment(ctx context.Context, pool *pgxpool.Pool, tableName string) (string, error) {
	query := `
		SELECT COALESCE(obj_description(c.oid), '') 
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = $1 AND n.nspname = $2
	`

	var comment string
	err := pool.QueryRow(ctx, query, tableName, i.schema).Scan(&comment)
	if err != nil && err != pgx.ErrNoRows {
		return "", err
	}

	return comment, nil
}

// getColumns gets all columns for a table
func (i *Introspector) getColumns(ctx context.Context, pool *pgxpool.Pool, tableName string) ([]Column, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			ordinal_position,
			COALESCE(col_description(pgc.oid, ordinal_position), '') as column_comment
		FROM information_schema.columns isc
		LEFT JOIN pg_class pgc ON pgc.relname = isc.table_name
		LEFT JOIN pg_namespace pgn ON pgn.oid = pgc.relnamespace AND pgn.nspname = isc.table_schema
		WHERE table_name = $1 AND table_schema = $2
		ORDER BY ordinal_position
	`

	rows, err := pool.Query(ctx, query, tableName, i.schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var isNullable string
		var defaultValue *string

		err := rows.Scan(&col.Name, &col.Type, &isNullable, &defaultValue, &col.Position, &col.Comment)
		if err != nil {
			return nil, err
		}

		col.IsNullable = isNullable == "YES"
		col.DefaultValue = defaultValue
		col.GoType = mapPostgresToGoType(col.Type, col.IsNullable)

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// getPrimaryKeys gets primary key columns for a table
func (i *Introspector) getPrimaryKeys(ctx context.Context, pool *pgxpool.Pool, tableName string) ([]string, error) {
	query := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		JOIN pg_class c ON c.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = $1 AND n.nspname = $2 AND i.indisprimary
		ORDER BY a.attnum
	`

	rows, err := pool.Query(ctx, query, tableName, i.schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var primaryKeys []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		primaryKeys = append(primaryKeys, columnName)
	}

	return primaryKeys, rows.Err()
}

// getIndexes gets all indexes for a table
func (i *Introspector) getIndexes(ctx context.Context, pool *pgxpool.Pool, tableName string) ([]Index, error) {
	query := `
		SELECT 
			i.relname as index_name,
			a.attname as column_name,
			ix.indisunique
		FROM pg_class t
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relname = $1 AND n.nspname = $2 AND t.relkind = 'r'
		ORDER BY i.relname, a.attnum
	`

	rows, err := pool.Query(ctx, query, tableName, i.schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexMap := make(map[string]*Index)
	for rows.Next() {
		var indexName, columnName string
		var isUnique bool

		err := rows.Scan(&indexName, &columnName, &isUnique)
		if err != nil {
			return nil, err
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &Index{
				Name:     indexName,
				Columns:  []string{columnName},
				IsUnique: isUnique,
			}
		}
	}

	var indexes []Index
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}

	return indexes, rows.Err()
}

// getForeignKeys gets all foreign keys for a table
func (i *Introspector) getForeignKeys(ctx context.Context, pool *pgxpool.Pool, tableName string) ([]ForeignKey, error) {
	query := `
		SELECT 
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name = $1 AND tc.table_schema = $2
	`

	rows, err := pool.Query(ctx, query, tableName, i.schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foreignKeys []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		err := rows.Scan(&fk.Name, &fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn)
		if err != nil {
			return nil, err
		}
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, rows.Err()
}

// mapPostgresToGoType maps PostgreSQL types to Go types
func mapPostgresToGoType(pgType string, isNullable bool) string {
	var goType string

	switch strings.ToLower(pgType) {
	case "integer", "int", "int4":
		goType = "int"
	case "bigint", "int8":
		goType = "int64"
	case "smallint", "int2":
		goType = "int16"
	case "serial", "serial4":
		goType = "int"
	case "bigserial", "serial8":
		goType = "int64"
	case "real", "float4":
		goType = "float32"
	case "double precision", "float8":
		goType = "float64"
	case "numeric", "decimal":
		goType = "decimal.Decimal"
	case "boolean", "bool":
		goType = "bool"
	case "character varying", "varchar", "character", "char", "text":
		goType = "string"
	case "date":
		goType = "time.Time"
	case "timestamp", "timestamp without time zone", "timestamp with time zone", "timestamptz":
		goType = "time.Time"
	case "time", "time without time zone", "time with time zone", "timetz":
		goType = "time.Time"
	case "uuid":
		goType = "uuid.UUID"
	case "json", "jsonb":
		goType = "json.RawMessage"
	case "bytea":
		goType = "[]byte"
	default:
		goType = "interface{}"
	}

	// Handle nullable types
	if isNullable && goType != "interface{}" {
		switch goType {
		case "int":
			return "*int"
		case "int64":
			return "*int64"
		case "int16":
			return "*int16"
		case "float32":
			return "*float32"
		case "float64":
			return "*float64"
		case "bool":
			return "*bool"
		case "string":
			return "*string"
		case "time.Time":
			return "*time.Time"
		default:
			return "*" + goType
		}
	}

	return goType
}
