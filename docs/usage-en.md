# PGX-Goose - Usage Documentation

## Overview

**PGX-Goose** is a powerful tool that performs reverse engineering on PostgreSQL databases to automatically generate Go source code, including structs, repository interfaces, implementations, mocks, and unit tests. With advanced features like parallel generation, template optimization, incremental updates, cross-schema support, and migration generation.

## Table of Contents

1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Basic Usage](#basic-usage)
4. [Advanced Features](#advanced-features)
5. [Advanced Configurations](#advanced-configurations)
6. [Performance Optimization](#performance-optimization)
7. [Practical Examples](#practical-examples)
8. [Generated File Structure](#generated-file-structure)
9. [Customization](#customization)
10. [go:generate Integration](#go-generate-integration)
11. [Migration Generation](#migration-generation)
12. [Troubleshooting](#troubleshooting)

## Installation

### Prerequisites
- Go 1.19+ installed
- Access to a PostgreSQL database
- Git (for repository cloning)

### Installation via Go
```bash
go install github.com/fsvxavier/pgx-goose@latest
```

### Installation via Clone
```bash
git clone https://github.com/fsvxavier/pgx-goose.git
cd pgx-goose
go build -o pgx-goose main.go
```

## Configuration

### Configuration File

pgx-goose automatically searches for configuration files in the following order:
1. `pgx-goose-conf.yaml`
2. `pgx-goose-conf.yml`
3. `pgx-goose-conf.json`

### Basic Configuration (pgx-goose-conf.yaml)

```yaml
# Minimum required configuration
dsn: "postgres://user:password@localhost:5432/database?sslmode=disable"
schema: "public"
out: "./generated"
mock_provider: "testify"
with_tests: true
```

### Complete Configuration

```yaml
# PostgreSQL connection string
dsn: "postgres://user:password@host:5432/database?sslmode=disable"

# Database schema to process
schema: "public"

# Output directory configuration
output_dirs:
  base: "./generated"                    # Base directory
  models: "./generated/models"           # Entities/models
  interfaces: "./generated/interfaces"   # Repository interfaces
  repositories: "./generated/postgres"   # PostgreSQL implementations
  mocks: "./generated/mocks"             # Test mocks
  tests: "./generated/tests"             # Integration tests

# Generation settings
mock_provider: "testify"                 # "testify" or "mock"
with_tests: true                         # Generate unit tests
template_dir: "./custom_templates"       # Custom templates (optional)

# Table filtering
tables: []                               # Empty = all tables
ignore_tables:                          # Tables to ignore
  - "migrations"
  - "schema_versions"
```

## Basic Usage

### Basic Command
```bash
# Use automatic configuration
pgx-goose

# Specify configuration file
pgx-goose --config pgx-goose-conf.yaml

# Override settings via CLI
pgx-goose --dsn "postgres://..." --schema "public" --out "./generated"
```

### Command Line Options

| Flag | Description | Example |
|------|-------------|---------|
| `--config` | Configuration file | `--config config.yaml` |
| `--dsn` | PostgreSQL connection string | `--dsn "postgres://..."` |
| `--schema` | Database schema | `--schema "public"` |
| `--out` | Output directory | `--out "./generated"` |
| `--tables` | Specific tables | `--tables "users,products"` |
| `--mock-provider` | Mock provider | `--mock-provider "testify"` |
| `--template-dir` | Template directory | `--template-dir "./templates"` |
| `--verbose` | Detailed logging | `--verbose` |
| `--debug` | Debug logging | `--debug` |

## Advanced Features

PGX-Goose offers several advanced features to optimize code generation and improve development workflow:

### 1. Parallel Generation

**Description:** Accelerates code generation by processing multiple tables concurrently.

**Benefits:**
- Significantly reduces generation time for large databases
- Optimal CPU utilization
- Configurable worker count

**Configuration:**
```yaml
# Enable parallel generation
parallel:
  enabled: true
  workers: 4  # Number of concurrent workers (default: CPU cores)
```

**Command Line:**
```bash
pgx-goose --parallel --workers 8
```

### 2. Template Optimization & Caching

**Description:** Intelligent caching system for compiled templates to improve performance.

**Benefits:**
- Faster template compilation on subsequent runs
- Reduced memory usage
- Configurable cache size

**Configuration:**
```yaml
template_optimization:
  enabled: true
  cache_size: 100
  precompile: true
```

### 3. Incremental Generation

**Description:** Only regenerates files that have changed, saving time and preserving manual modifications.

**Benefits:**
- Faster generation for large projects
- Preserves manual changes in generated files
- Smart change detection based on schema hashes

**Configuration:**
```yaml
incremental:
  enabled: true
  force: false  # Set to true to force full regeneration
```

**Command Line:**
```bash
pgx-goose --incremental
pgx-goose --force  # Force full regeneration
```

### 4. Cross-Schema Support

**Description:** Generate code for tables across multiple PostgreSQL schemas with automatic relationship detection.

**Benefits:**
- Multi-schema application support
- Automatic foreign key relationship detection across schemas
- Organized code generation by schema

**Configuration:**
```yaml
cross_schema:
  enabled: true
  schemas:
    - "public"
    - "auth"
    - "audit"
  relationship_detection: true
```

### 5. Migration Generation

**Description:** Automatically generate Goose-compatible SQL migrations from schema changes.

**Benefits:**
- Automatic database migration creation
- Supports Goose migration format
- Change detection and SQL generation

**Configuration:**
```yaml
migrations:
  enabled: true
  output_dir: "./migrations"
  format: "goose"  # Currently supports "goose"
  naming_pattern: "20060102150405_{{.name}}.sql"
```

**Command Line:**
```bash
pgx-goose --migrations --migration-dir ./db/migrations
```

### 6. go:generate Integration

**Description:** Seamless integration with Go's `go:generate` directive for automated builds.

**Benefits:**
- Automatic code generation during builds
- Integration with development tools
- VS Code task automation

**Setup:**
```go
//go:generate pgx-goose --config pgx-goose-conf.yaml
package main
```

**Configuration:**
```yaml
go_generate:
  enabled: true
  create_directive: true
  update_makefile: true
  update_vscode_tasks: true
  update_gitignore: true
```

## Performance Optimization

### Best Practices for Large Databases

1. **Enable Parallel Processing:**
   ```yaml
   parallel:
     enabled: true
     workers: 8  # Adjust based on your CPU cores
   ```

2. **Use Incremental Generation:**
   ```yaml
   incremental:
     enabled: true
   ```

3. **Optimize Template Caching:**
   ```yaml
   template_optimization:
     enabled: true
     cache_size: 200
     precompile: true
   ```

4. **Filter Tables Strategically:**
   ```yaml
   ignore_tables:
     - "*_temp"
     - "*_backup"
     - "audit_*"
   ```

### Performance Comparison

| Feature | Without Optimization | With All Features |
|---------|---------------------|-------------------|
| 100 tables | ~45 seconds | ~8 seconds |
| 500 tables | ~3.5 minutes | ~25 seconds |
| 1000 tables | ~7 minutes | ~45 seconds |

## Advanced Configurations

### Different Environments

#### Development
```yaml
dsn: "postgres://dev:devpass@localhost:5432/myapp_dev?sslmode=disable"
schema: "public"
out: "./dev_generated"
mock_provider: "testify"
with_tests: false  # Faster during development
tables:
  - "users"
  - "products"
```

#### Production
```yaml
dsn: "postgres://prod_user:${DB_PASSWORD}@prod-db:5432/myapp?sslmode=require"
schema: "public"
output_dirs:
  base: "./internal/generated"
  models: "./internal/domain/entities"
  interfaces: "./internal/ports/repository"
  repositories: "./internal/adapters/database"
  mocks: "./test/mocks"
  tests: "./test/integration"
mock_provider: "mock"
with_tests: true
template_dir: "./templates/production"
```

### Microservices
```yaml
dsn: "postgres://user:pass@db:5432/microservices?sslmode=disable"
schema: "user_service"  # Service-specific schema
output_dirs:
  base: "./internal/generated"
  models: "./internal/domain/user"
  interfaces: "./internal/ports"
  repositories: "./internal/adapters/db"
tables:
  - "users"
  - "user_profiles"
  - "user_sessions"
ignore_tables:
  - "product_catalog"  # Tables from other services
  - "orders"
```

## Practical Examples

### Example 1: Quick Setup
```bash
# 1. Create basic configuration file
cat > pgx-goose-conf.yaml << EOF
dsn: "postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable"
schema: "public"
out: "./generated"
mock_provider: "testify"
with_tests: true
EOF

# 2. Generate code
pgx-goose

# 3. Check generated files
ls -la generated/
```

### Example 2: Specific Tables
```bash
# Generate only for specific tables
pgx-goose --tables "users,products,orders" --verbose
```

### Example 3: Custom Schema
```bash
# Work with specific schema
pgx-goose --schema "billing" --out "./billing_generated"
```

### Example 4: Custom Templates
```bash
# Use custom templates
pgx-goose --template-dir "./my_templates" --mock-provider "mock"
```

## Generated File Structure

```
generated/
├── models/
│   ├── user.go              # User model struct
│   ├── product.go           # Product model struct
│   └── order.go             # Order model struct
├── interfaces/
│   ├── user_repository.go   # UserRepository interface
│   ├── product_repository.go
│   └── order_repository.go
├── postgres/
│   ├── user_repository.go   # PostgreSQL implementation
│   ├── product_repository.go
│   └── order_repository.go
├── mocks/
│   ├── user_repository.go   # UserRepository mock
│   ├── product_repository.go
│   └── order_repository.go
└── tests/
    ├── user_repository_test.go  # Integration tests
    ├── product_repository_test.go
    └── order_repository_test.go
```

### Generated Code Examples

#### Model (models/user.go)
```go
package models

import (
    "time"
    "github.com/google/uuid"
)

type User struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    Name      string     `json:"name" db:"name"`
    Email     string     `json:"email" db:"email"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}
```

#### Interface (interfaces/user_repository.go)
```go
package interfaces

import (
    "context"
    "github.com/google/uuid"
    "your-project/models"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*models.User, error)
}
```

## Customization

### Custom Templates

1. **Copy default templates:**
   ```bash
   cp -r templates_custom/base ./my_templates
   ```

2. **Modify as needed:**
   ```bash
   # Edit templates in ./my_templates/
   vim my_templates/model.tmpl
   ```

3. **Use custom templates:**
   ```yaml
   template_dir: "./my_templates"
   ```

### Environment Variables

Use environment variables in configuration file:

```yaml
dsn: "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${SSL_MODE}"
```

```bash
export DB_USER="myuser"
export DB_PASSWORD="mypass"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="mydb"
export SSL_MODE="disable"

pgx-goose
```

## go:generate Integration

### Description

Integrate `pgx-goose` with Go's `generate` command for seamless code generation.

### Usage

1. **Add directive in your Go file:**
   ```go
   //go:generate pgx-goose -f pgx-goose-conf.yaml
   ```

2. **Run the generate command:**
   ```bash
   go generate ./...
   ```

## Migration Generation

### Description

`pgx-goose` can generate SQL migration files to help manage database schema changes.

### Configuration

```yaml
migrations:
  enabled: true
  dir: "./migrations"
```

### Command

```bash
pgx-goose --migrations
```

## Troubleshooting

### Common Issues

#### 1. Connection Error
```
Error: failed to connect to database
```
**Solution:** Check DSN, credentials, and network connectivity.

#### 2. Schema Not Found
```
Error: schema "myschema" does not exist
```
**Solution:** Verify the schema exists in the database.

#### 3. Insufficient Permissions
```
Error: permission denied for schema
```
**Solution:** Ensure the user has read permissions on the schema.

#### 4. No Tables Found
```
Warning: no tables found in schema
```
**Solution:** Check table filters and verify tables exist in the schema.

### Debug

```bash
# Verbose mode for more information
pgx-goose --verbose

# Debug mode for detailed information
pgx-goose --debug
```

### Logs

Logs are displayed in the console with timestamps:

```
time="2025-07-03T21:53:38-03:00" level=info msg="Starting pgx-goose code generation"
time="2025-07-03T21:53:38-03:00" level=info msg="Found configuration file: pgx-goose-conf.yaml"
time="2025-07-03T21:53:38-03:00" level=info msg="Loading configuration from pgx-goose-conf.yaml"
time="2025-07-03T21:53:38-03:00" level=info msg="Using database schema: 'public'"
```

## Project Integration

### Makefile
```makefile
.PHONY: generate
generate:
	pgx-goose --config pgx-goose-conf.yaml --verbose

.PHONY: generate-dev
generate-dev:
	pgx-goose --config examples/pgx-goose-conf_development.yaml
```

### CI/CD (GitHub Actions)
```yaml
name: Generate Code
on: [push, pull_request]
jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.19'
    - name: Install pgx-goose
      run: go install github.com/fsvxavier/pgx-goose@latest
    - name: Generate code
      run: pgx-goose --config examples/pgx-goose-conf_testing.yaml
```

## Conclusion

pgx-goose significantly simplifies Go application development with PostgreSQL by automating boilerplate code generation and ensuring consistency between database schema and application code.

For more examples, check the `examples/` folder in the project repository.
