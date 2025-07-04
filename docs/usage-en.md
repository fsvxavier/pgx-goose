# PGX-Goose - Usage Documentation

## Overview

**PGX-Goose** is a powerful tool that performs reverse engineering on PostgreSQL databases to automatically generate Go source code, including structs, repository interfaces, implementations, mocks, and unit tests.

## Table of Contents

1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Basic Usage](#basic-usage)
4. [Advanced Configurations](#advanced-configurations)
5. [Practical Examples](#practical-examples)
6. [Generated File Structure](#generated-file-structure)
7. [Customization](#customization)
8. [Troubleshooting](#troubleshooting)

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
