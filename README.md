# PGX-Goose

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**PGX-Goose** is a PostgreSQL reverse engineering tool that automatically generates idiomatic Go code including structs, repository interfaces, implementations, mocks, and unit tests. Supports multiple schemas for complex enterprise architectures.

> 🇺🇸 **English version (current)** | 🇧🇷 **[Versão em português disponível](README-pt-br.md)** | 🇪🇸 **[Versión en español disponible](README-es.md)**

## 📋 Table of Contents

- [🚀 Features](#-features)
- [📦 Installation](#-installation)
- [⚡ Quick Start](#-quick-start)
- [⚙️ Configuration](#️-configuration)
- [📁 Generated Structure](#-generated-structure)
- [🎨 Templates](#-templates)
- [🔧 CLI Reference](#-cli-reference)
- [💡 Usage Examples](#-usage-examples)
- [🤝 Contributing](#-contributing)
- [ Documentation](#-documentation)

## 🚀 Features

- **🔍 Complete Analysis**: Introspects PostgreSQL schemas (tables, columns, types, PKs, indexes, relationships)
- **🏢 Multi-Schema**: Support for custom schemas for enterprise architectures
- **🤖 Automatic Generation**: Creates structs, interfaces, implementations, mocks and tests
- **📂 Flexible Directories**: Customizable output directory configuration
- **🎨 Customizable Templates**: Custom Go templates + optimized PostgreSQL templates (including simple variations)
- **🧪 Mock Providers**: Support for `testify/mock`, `mock` and `gomock`
- **🎯 Clean Architecture**: Code following Clean Architecture and SOLID principles
- **⚡ Advanced Operations**: Transactions, batch operations and soft delete
- **🔧 Robust CLI**: Complete command line interface with validation and configurable logging
- **📝 Flexible Configuration**: YAML/JSON support with hierarchical precedence

### 🚀 Advanced Features

- **⚡ Parallel Generation**: Multi-worker concurrent processing for improved performance
- **🎯 Incremental Generation**: Smart change detection to regenerate only modified files
- **📦 Template Optimization**: Intelligent caching system for compiled templates
- **🔄 Cross-Schema Support**: Generate code across multiple PostgreSQL schemas with relationship detection
- **🗄️ Migration Generation**: Automatic Goose-compatible SQL migration creation
- **🛠️ go:generate Integration**: Seamless integration with Go's build system

## 📦 Installation

### Via go install (Recommended)

```bash
go install github.com/fsvxavier/pgx-goose@latest
```

### Local build

```bash
git clone https://github.com/fsvxavier/nexs-lib.git
cd pgx-goose
go build -o pgx-goose .
./pgx-goose --help
```

## 📚 Documentation

Complete documentation available in multiple languages:

- 🇧🇷 **[Português (Brasil)](docs/usage-pt-br.md)** - Complete documentation in Brazilian Portuguese
- 🇺🇸 **[English](docs/usage-en.md)** - Complete documentation in English  
- 🇪🇸 **[Español](docs/usage-es.md)** - Complete documentation in Spanish
- 📋 **[Quick Reference](docs/quick-reference.md)** - Quick reference for commands and configurations

### What's covered in the documentation:
- Detailed installation and prerequisites
- Complete configuration (YAML/JSON)
- Basic and advanced usage
- Practical examples for different scenarios
- Generated file structure
- Template customization
- Troubleshooting and problem solving
- Project integration (Makefile, CI/CD)

### Configuration Examples
See the [examples/](examples/) folder for:
- Basic and advanced configurations
- Environment-specific setups (dev, prod, testing)
- Microservice configurations
- Table filtering examples

## ⚡ Quick Start

### 1. Simple Command
```bash
# Generate code for all tables
pgx-goose --dsn "postgres://user:pass@localhost:5432/mydb"
```

### 2. With YAML Configuration
```yaml
# pgx-goose-conf.yaml
dsn: "postgres://user:pass@localhost:5432/mydb"
schema: "public"
out: "./generated"
template_dir: "./templates_postgresql"
mock_provider: "testify"
with_tests: true
```

```bash
pgx-goose --config pgx-goose-conf.yaml
```

### 3. Common Commands
```bash
# Specific tables
pgx-goose --dsn "..." --tables "users,orders,products"

# Custom schema
pgx-goose --dsn "..." --schema "inventory" --out "./inventory-gen"

# Optimized PostgreSQL templates
pgx-goose --config pgx-goose-conf.yaml --template-dir "./templates_postgresql"
```

## ⚙️ Configuration

### Configuration File

#### pgx-goose-conf.yaml (Recommended)
```yaml
# Connection
dsn: "postgres://user:pass@localhost:5432/db?sslmode=disable"
schema: "public"  # Custom schema (default: "public")

# Output directories
output_dirs:
  base: "./generated"                       # Base directory (default: ./pgx-goose)
  models: "./internal/domain/entities"      # Structs
  interfaces: "./internal/ports"            # Interfaces
  repositories: "./internal/adapters/db"    # Implementations
  mocks: "./tests/mocks"                    # Mocks
  tests: "./tests/integration"              # Tests

# Table filters
tables: []                                  # [] = all, or ["users", "orders"] 
ignore_tables:                             # Tables to ignore
  - "schema_migrations"       # Rails/Laravel migrations
  - "ar_internal_metadata"    # Rails metadata
  - "goose_db_version"        # Goose migrations
  - "migrations"              # Generic migrations
  - "audit_logs"              # Audit/log tables
  - "sessions"                # Session data

# Generation options
template_dir: "./templates_postgresql"  # Optimized templates
mock_provider: "testify"                # "testify", "mock", or "gomock"  
with_tests: true                        # Generate tests (default: true)
```

#### pgx-goose-conf.json (Alternative)
```json
{
  "dsn": "postgres://user:pass@localhost:5432/db",
  "schema": "public",
  "output_dirs": {
    "base": "./generated",
    "models": "./models",
    "interfaces": "./repositories/interfaces", 
    "repositories": "./repositories/postgres",
    "mocks": "./mocks",
    "tests": "./tests"
  },
  "tables": [],
  "ignore_tables": ["migrations", "logs", "sessions"],
  "template_dir": "./templates_postgresql",
  "mock_provider": "testify",
  "with_tests": true
}
```

### Detailed Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dsn` | string | **required** | PostgreSQL connection string |
| `schema` | string | `"public"` | Database schema to introspect |
| `output_dirs.base` | string | `"./pgx-goose"` | Base output directory |
| `output_dirs.models` | string | `"{base}/models"` | Directory for structs |
| `output_dirs.interfaces` | string | `"{base}/repository/interfaces"` | Directory for interfaces |
| `output_dirs.repositories` | string | `"{base}/repository/postgres"` | Directory for implementations |
| `output_dirs.mocks` | string | `"{base}/mocks"` | Directory for mocks |
| `output_dirs.tests` | string | `"{base}/tests"` | Directory for tests |
| `tables` | []string | `[]` (all) | List of specific tables |
| `ignore_tables` | []string | `[]` | List of tables to ignore |
| `template_dir` | string | `""` (built-in) | Custom templates directory |
| `mock_provider` | string | `"testify"` | Mock provider: `testify`, `mock`, `gomock` |
| `with_tests` | bool | `true` | Generate test files |

### Validation and Rules

1. **Required DSN**: The `dsn` field is always required
2. **Table conflicts**: Not allowed to specify the same table in `tables` and `ignore_tables`
3. **Valid mock providers**: Only `testify`, `mock` and `gomock` are accepted
4. **Directories**: If not specified, use defaults relative to `base`

### Configuration Precedence

Configuration follows a precedence hierarchy (highest to lowest):

1. **CLI flags** (highest precedence)
2. **Configuration file** (`--config`)
3. **Default values** (lowest precedence)

```bash
# CLI overrides any value from configuration file
pgx-goose --config pgx-goose-conf.yaml --schema "billing" --mock-provider "gomock"
```

### Table Filtering

#### Inclusive Mode (Specific Tables)
```yaml
tables: ["users", "orders", "products"]  # Only these tables
ignore_tables: []                        # Ignore list must be empty
```

#### Exclusive Mode (All Except...)
```yaml
tables: []  # Empty list = all tables
ignore_tables: 
  - "schema_migrations"      # Rails/Laravel
  - "ar_internal_metadata"   # ActiveRecord
  - "goose_db_version"       # Goose migrations
  - "audit_logs"             # Audit logs
  - "sessions"               # User sessions
```

#### Conflict Validation
```yaml
# ❌ ERROR: Conflict detected - table in both lists
tables: ["users", "orders"]
ignore_tables: ["users"]  # users appears in both lists

# ✅ OK: No conflicts
tables: ["users", "orders"] 
ignore_tables: []
```

### Validation Rules

The system applies the following validations before execution:

| Validation | Description | Error |
|------------|-------------|-------|
| **Required DSN** | `dsn` field must be present | `DSN is required` |
| **Valid mock provider** | Must be `testify`, `mock` or `gomock` | `invalid mock provider` |
| **Table conflicts** | Table cannot be in `tables` AND `ignore_tables` | `conflicting table configuration` |
| **Config file** | If specified, must exist and be valid | `failed to read config file` |
| **Config format** | Must be `.yaml`, `.yml` or `.json` | `unsupported config file format` |

### Logging Configuration

```bash
# Default logging (warnings/errors only)
pgx-goose --config pgx-goose-conf.yaml

# Verbose (info + warnings + errors)
pgx-goose --config pgx-goose-conf.yaml --verbose

# Debug (everything)
pgx-goose --config pgx-goose-conf.yaml --debug
```

## 📁 Generated Structure

### Default Structure
```
generated/
├── models/                 # Entity structs
│   ├── user.go
│   └── product.go
├── repository/
│   ├── interfaces/         # Repository interfaces
│   │   ├── user_repository.go
│   │   └── product_repository.go
│   └── postgres/           # PostgreSQL implementations
│       ├── user_repository.go
│       └── product_repository.go
├── mocks/                  # Test mocks
│   ├── mock_user_repository.go
│   └── mock_product_repository.go
└── tests/                  # Unit/integration tests
    ├── user_repository_test.go
    └── product_repository_test.go
```

### Custom Directory Structure
```
internal/
├── domain/entities/        # Models
├── ports/                  # Interfaces  
└── adapters/postgres/      # Implementations
tests/
├── mocks/                  # Mocks
└── integration/            # Tests
```

### Types of Generated Files

| Type | Description | Content |
|------|-------------|---------|
| **Models** | Entity structs | JSON/DB tags, validation, utility methods |
| **Interfaces** | Repository contracts | CRUD, transactions, batch operations |
| **Repositories** | PostgreSQL implementations | Connection pools, prepared statements |
| **Mocks** | Testify/GoMock | Expectation methods, assertions |
| **Tests** | Integration tests | Setup/teardown, benchmarks, testcontainers |

## 🎨 Templates

### Available Templates

#### 1. Default Templates (`./templates/`)
- Generic templates for any Go project
- Basic pgx compatibility

#### 2. PostgreSQL Templates (`./templates_postgresql/`)
- **Recommended** - Optimized for `nexs-lib`
- Transaction and batch operation support
- Advanced struct methods

#### 3. Template Variations

Each template set has two variations:

| Template | Default | Simple (`*_simple.tmpl`) |
|----------|---------|-------------------------|
| **Model** | Complete struct with utility methods | Basic struct only |
| **Repository** | Complete interface/implementation | Basic CRUD operations only |
| **Mock** | Complete mock with all methods | Simplified mock |
| **Test** | Comprehensive tests with benchmarks | Basic unit tests |

### Using PostgreSQL Templates
```bash
pgx-goose --template-dir "./templates_postgresql" --config pgx-goose-conf.yaml
```

### Custom Templates

Create a directory with:
```
my_templates/
├── model.tmpl                  # Structs
├── repository_interface.tmpl   # Interfaces
├── repository_postgres.tmpl    # Implementations
├── mock_testify.tmpl          # Testify mocks
├── mock_gomock.tmpl           # GoMock mocks
└── test.tmpl                  # Tests
```

**Example model.tmpl:**
```go
package {{.Package}}

import "time"

// {{.StructName}} represents {{.Table.Comment}}
type {{.StructName}} struct {
{{- range .Table.Columns}}
    {{toPascalCase .Name}} {{.GoType}} `json:"{{.Name}}" db:"{{.Name}}"`
{{- end}}
}

func (e *{{.StructName}}) TableName() string {
    return "{{.Table.Name}}"
}
```

## 🔧 CLI Reference

### Main Flags

| Flag | Description | Values | Default |
|------|-------------|--------|---------|
| `--dsn` | PostgreSQL connection string | `postgres://user:pass@host:port/db` | **required** |
| `--schema` | Database schema | `public`, `inventory`, `billing` | `public` |
| `--config` | Configuration file | `pgx-goose-conf.yaml`, `pgx-goose-conf.json` | - |
| `--out` | Output directory | `./generated` | `./pgx-goose` |
| `--tables` | Specific tables (CSV) | `users,orders,products` | all |
| `--template-dir` | Templates directory | `./templates_postgresql` | built-in |
| `--mock-provider` | Mock provider | `testify`, `mock`, `gomock` | `testify` |
| `--with-tests` | Generate tests | `true`, `false` | `true` |

### Specific Directory Flags

| Flag | Directory | Example |
|------|-----------|---------|
| `--models-dir` | Models/structs | `./internal/domain/entities` |
| `--interfaces-dir` | Interfaces | `./internal/ports` |
| `--repos-dir` | Implementations | `./internal/adapters/postgres` |
| `--mocks-dir` | Mocks | `./tests/mocks` |
| `--tests-dir` | Tests | `./tests/integration` |

### Configuration and Logging Flags

| Flag | Description | Usage |
|------|-------------|-------|
| `--json` | Use JSON format for configuration | To prefer .json over .yaml |
| `--yaml` | Use YAML format for configuration | Default, explicit |
| `--verbose` | Verbose logging (INFO level) | Execution debugging |
| `--debug` | Debug logging (DEBUG level) | Complete debugging |

### Command Examples

```bash
# Basic
pgx-goose --dsn "postgres://user:pass@localhost:5432/db"

# Custom schema + specific tables
pgx-goose --dsn "..." --schema "billing" --tables "invoices,payments"

# Complete configuration with logging
pgx-goose --config pgx-goose-conf.yaml --template-dir "./templates_postgresql" --verbose

# Ignore specific tables
pgx-goose --dsn "..." --ignore-tables "migrations,logs,sessions"

# Modular organization with custom directories
pgx-goose --dsn "..." --tables "users" \
  --models-dir "./modules/user/entity" \
  --interfaces-dir "./modules/user/repository"

# Enterprise multi-schema
pgx-goose --schema "inventory" --out "./modules/inventory/generated"
pgx-goose --schema "billing" --out "./modules/billing/generated"

# Custom mock provider
pgx-goose --config pgx-goose-conf.yaml --mock-provider "gomock" --debug
```

## 💡 Usage Examples

### 1. Simple E-commerce Project

**SQL Schema:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INTEGER DEFAULT 0
);
```

**Configuration:**
```yaml
# pgx-goose-conf.yaml
dsn: "postgres://admin:pass@localhost:5432/ecommerce"
output_dirs:
  models: "./internal/domain"
  interfaces: "./internal/ports"
  repositories: "./internal/adapters/postgres"
template_dir: "./templates_postgresql"
mock_provider: "testify"
```

**Generate code:**
```bash
pgx-goose --config pgx-goose-conf.yaml
```

### 2. Enterprise Multi-Schema Architecture

```bash
# User schema
pgx-goose --schema "users" --out "./modules/users/generated"

# Inventory schema
pgx-goose --schema "inventory" --out "./modules/inventory/generated"

# Billing schema
pgx-goose --schema "billing" --out "./modules/billing/generated"
```

### 3. Generated Code - User Example

**Generated Model (`models/user.go`):**
```go
type User struct {
    ID        int64     `json:"id" db:"id"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    Name      string    `json:"name" db:"name" validate:"required"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (u *User) TableName() string { return "users" }
func (u *User) Validate() error { /* validation */ }
func (u *User) Clone() *User { /* safe cloning */ }
```

**Generated Interface (`interfaces/user_repository.go`):**
```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id int64) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id int64) error
    
    // Transactions
    CreateTx(ctx context.Context, tx common.ITransaction, user *models.User) error
    
    // Batch operations
    CreateBatch(ctx context.Context, users []*models.User) error
    
    // Specific searches
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}
```

### 4. Using Generated Code

```go
package main

import (
    "context"
    "your-project/internal/domain"
    "your-project/internal/adapters/postgres"
    "github.com/fsvxavier/nexs-lib/db/postgresql"
)

func main() {
    // Configure PostgreSQL pool
    pool, _ := postgresql.NewPool(postgresql.Config{
        Host:     "localhost",
        Database: "ecommerce",
        Username: "admin",
        Password: "password",
    })
    defer pool.Close()
    
    // Use generated repository
    userRepo := postgres.NewUserRepository(pool)
    
    // Create user
    user := &domain.User{
        Email: "john@example.com",
        Name:  "John Doe",
    }
    
    err := userRepo.Create(context.Background(), user)
    if err != nil {
        panic(err)
    }
    
    // Find by email
    found, _ := userRepo.FindByEmail(context.Background(), "john@example.com")
    fmt.Printf("User found: %+v\n", found)
}
```

### 5. Testing with Mocks

```go
func TestUserService_CreateUser(t *testing.T) {
    mockRepo := &mocks.MockUserRepository{}
    service := NewUserService(mockRepo)
    
    user := &domain.User{Email: "test@example.com", Name: "Test"}
    mockRepo.On("Create", mock.Anything, user).Return(nil)
    
    err := service.CreateUser(context.Background(), user)
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

## 🤝 Contributing

### How to Contribute

1. **Fork** the project
2. **Create a branch** (`git checkout -b feature/NewFeature`)
3. **Commit** your changes (`git commit -m 'Add: new feature'`)
4. **Push** to the branch (`git push origin feature/NewFeature`)
5. **Open a Pull Request**

### Local Development

```bash
# Clone and setup
git clone https://github.com/fsvxavier/nexs-lib.git
cd pgx-goose
go mod download

# Tests
go test ./...

# Build
go build -o pgx-goose .
./pgx-goose --help
```

### Project Structure

```
pgx-goose/
├── cmd/                    # CLI commands (Cobra)
├── internal/
│   ├── config/            # Configuration
│   ├── generator/         # Code generation
│   └── introspector/      # PostgreSQL introspection
├── templates/             # Default templates
├── templates_postgresql/  # Optimized templates
├── examples/              # Configuration examples
└── docs/                  # Additional documentation
```

### Guidelines

- **Tests**: Every new feature must have tests
- **Documentation**: Update README.md for new features
- **Templates**: Maintain compatibility with existing templates
- **Logs**: Use slog for structured logging

---

## 📄 License

Licensed under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- [pgx](https://github.com/jackc/pgx) - High-performance PostgreSQL driver
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [testify](https://github.com/stretchr/testify) - Testing framework
- [testcontainers](https://github.com/testcontainers/testcontainers-go) - Integration testing

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- **Discussions**: [GitHub Discussions](https://github.com/fsvxavier/nexs-lib/discussions)

---

**PGX-Goose** - Transforming your PostgreSQL into idiomatic Go code! 🚀
