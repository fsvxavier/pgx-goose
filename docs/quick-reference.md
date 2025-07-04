# PGX-Goose Quick Reference

Quick reference guide for the most common PGX-Goose commands and configurations.

## Quick Commands

```bash
# Basic usage (auto-detect config)
pgx-goose

# With specific config file
pgx-goose --config pgx-goose-conf.yaml

# Override DSN
pgx-goose --dsn "postgres://user:pass@host:5432/db"

# Specific tables only
pgx-goose --tables "users,products,orders"

# With verbose logging
pgx-goose --verbose

# With debug logging
pgx-goose --debug
```

## Configuration Templates

### Minimal Config
```yaml
dsn: "postgres://user:pass@host:5432/db?sslmode=disable"
schema: "public"
out: "./generated"
```

### Complete Config
```yaml
dsn: "postgres://user:pass@host:5432/db?sslmode=disable"
schema: "public"
output_dirs:
  base: "./generated"
  models: "./generated/models"
  interfaces: "./generated/interfaces"
  repositories: "./generated/postgres"
  mocks: "./generated/mocks"
  tests: "./generated/tests"
mock_provider: "testify"
with_tests: true
tables: []
ignore_tables:
  - "migrations"
  - "schema_versions"
```

## Common Use Cases

### Development Setup
```bash
# Quick setup for development
echo 'dsn: "postgres://dev:devpass@localhost:5432/myapp_dev?sslmode=disable"
schema: "public"
out: "./dev_generated"
mock_provider: "testify"
with_tests: false' > pgx-goose-conf.yaml

pgx-goose
```

### Production Generation
```bash
pgx-goose --config examples/pgx-goose-conf_production.yaml --verbose
```

### Microservice
```bash
pgx-goose --schema "user_service" --tables "users,user_profiles" --out "./internal/generated"
```

### Custom Templates
```bash
pgx-goose --template-dir "./custom_templates" --mock-provider "mock"
```

## Troubleshooting

### Check Connection
```bash
# Test with verbose output
pgx-goose --dsn "your-dsn-here" --verbose --tables "non_existent_table"
```

### Debug Schema Issues
```bash
# List available schemas (if you have psql)
psql "your-dsn" -c "\dn"

# Check tables in schema
psql "your-dsn" -c "\dt your_schema.*"
```

### Permission Issues
```bash
# Test with a simple query
psql "your-dsn" -c "SELECT current_user, current_database(), current_schema();"
```

## Environment Variables

```bash
# Set common environment variables
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="myuser"
export DB_PASSWORD="mypass"
export DB_NAME="mydb"

# Use in config
# dsn: "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
```

## Integration Examples

### Makefile
```makefile
generate:
	pgx-goose --verbose

generate-dev:
	pgx-goose --config examples/pgx-goose-conf_development.yaml

generate-prod:
	pgx-goose --config examples/pgx-goose-conf_production.yaml
```

### Docker Compose
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
  
  generate:
    build: .
    depends_on:
      - postgres
    command: pgx-goose --dsn "postgres://user:password@postgres:5432/myapp?sslmode=disable"
    volumes:
      - ./generated:/app/generated
```

## File Structure Output

```
generated/
├── models/           # Database entities
├── interfaces/       # Repository interfaces  
├── postgres/         # PostgreSQL implementations
├── mocks/           # Test mocks
└── tests/           # Integration tests
```

## Quick Tips

- **Auto-detection**: Place `pgx-goose-conf.yaml` in current directory
- **Multiple environments**: Use different config files
- **Large databases**: Use `ignore_tables` to exclude system tables
- **Clean architecture**: Use `output_dirs` for custom organization
- **CI/CD**: Use `examples/pgx-goose-conf_testing.yaml` as base
- **Debugging**: Always use `--verbose` or `--debug` when troubleshooting
