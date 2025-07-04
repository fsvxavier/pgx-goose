# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-03

### 🎯 Main Features Implemented

#### ✅ Configurable Output Directories (NEW FEATURE)
- Added `output_dirs` configuration for separate directories
- Support for specific CLI flags (`--models-dir`, `--interfaces-dir`, etc.)
- Backward compatibility with legacy `OutputDir` configuration
- Precedence: CLI flags > config file > defaults
- Support for 5 different architectures (Hexagonal, Clean, DDD, Modular, Monorepo)

#### ✅ Optimized PostgreSQL Templates (NEW FEATURE)
- Specialized templates in `templates_postgresql/` folder
- Integration with `db/postgresql` provider from isis-golang-lib
- Support for transactional and batch operations
- Advanced methods in entities (TableName, Clone, Validate, etc.)
- Soft delete when applicable

#### ✅ Unified Documentation (NEW FEATURE)
- Complete README.md with ~1,315 lines
- 17 main sections with 15+ complete examples
- Integration of 6 documentation files
- Detailed index for navigation
- Use cases by architecture

### Added

#### Core Features
- ✅ Complete CLI tool based on Cobra
- ✅ Complete PostgreSQL schema introspection
- ✅ Automatic Go struct generation from tables
- ✅ Repository interface generation with complete CRUD
- ✅ PostgreSQL implementations using pgx/v5
- ✅ Support for two mock providers: testify and gomock
- ✅ Automatic unit test generation
- ✅ Customizable template system using Go Templates

#### Configuration & CLI
- ✅ Support for YAML and JSON configuration files
- ✅ Comprehensive command line flags
- ✅ Configurable logging system (debug, verbose, info, warn, error)
- ✅ Robust configuration validation

#### Database Support
- ✅ Complete introspection of tables, columns, types
- ✅ Support for primary keys, indexes, and foreign keys
- ✅ Automatic PostgreSQL → Go type mapping
- ✅ Support for nullable types with pointers
- ✅ Preserved table and column comments

#### Code Generation
- ✅ Embedded templates for all file types
- ✅ Organized and idiomatic project structure
- ✅ Support for custom templates via custom directory
- ✅ Clean code generation following Go conventions
- ✅ Support for relationships between tables

#### Testing & Quality
- ✅ Comprehensive unit tests
- ✅ Automatic mocks for all interfaces
- ✅ Generated tests with success and error scenarios
- ✅ Integration with testify/assert and testify/mock
- ✅ Support for gomock for projects that prefer it

### Technical Features

#### Architecture
- Clean Architecture with separation of concerns
- Interface-oriented design
- Dependency injection
- Modular and extensible structure

#### Supported Types
| PostgreSQL | Go |
|------------|-----|
| integer, int, int4 | int |
| bigint, int8 | int64 |
| smallint, int2 | int16 |
| real, float4 | float32 |
| double precision, float8 | float64 |
| numeric, decimal | decimal.Decimal |
| boolean, bool | bool |
| varchar, text, char | string |
| date, timestamp | time.Time |
| uuid | uuid.UUID |
| json, jsonb | json.RawMessage |
| bytea | []byte |

#### Generated Project Structure
```
output_dir/
├── models/                     # Table structs
├── repository/
│   ├── interfaces/             # Repository interfaces
│   └── postgres/               # PostgreSQL implementations
├── mocks/                      # Mocks for testing
└── tests/                      # Unit tests
```

#### Available CLI Flags
- `--dsn`: PostgreSQL connection string
- `--out`: Output directory
- `--tables`: List of specific tables
- `--config`: Configuration file
- `--template-dir`: Custom templates
- `--mock-provider`: Mock provider (testify/mock)
- `--with-tests`: Test generation
- `--verbose`: Verbose logging
- `--debug`: Debug logging

### Project Files

#### Documentation
- ✅ Complete README.md with examples
- ✅ EXAMPLES.md with detailed use cases
- ✅ Configuration templates (YAML/JSON)
- ✅ Demo scripts (Bash/PowerShell)

#### Build & Development
- ✅ Makefile with useful targets
- ✅ go.mod with all dependencies
- ✅ Appropriate .gitignore
- ✅ MIT License

#### Scripts & Tools
- ✅ demo.sh for Linux/macOS
- ✅ demo.ps1 for Windows
- ✅ Example SQL schema
- ✅ Example configurations

### Dependencies

#### Runtime
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/spf13/cobra` - CLI framework
- `log/slog` - Structured logging (native Go 1.21+)
- `gopkg.in/yaml.v3` - YAML parser

#### Testing & Mocking
- `github.com/stretchr/testify` - Testing framework
- `go.uber.org/mock` - Gomock for mocks
- `github.com/google/uuid` - UUID support
- `github.com/shopspring/decimal` - Decimal types

### Usage Notes

#### Installation
```bash
go install github.com/fsvxavier/pgx-goose@latest
```

#### Basic Usage
```bash
pgx-goose --dsn "postgres://user:pass@localhost:5432/db" --out ./generated
```

#### With Configuration
```bash
pgx-goose --config pgx-goose-conf.yaml --verbose
```

### Acknowledgments

This project was inspired by:
- [xo/dbtpl](https://github.com/xo/dbtpl)
- [go-gorm/gen](https://github.com/go-gorm/gen)
- Clean Architecture principles
- SOLID and DDD patterns

### Contributing

To contribute to the project:
1. Fork the repository
2. Create a branch for your feature
3. Implement with tests
4. Run tests and linting
5. Submit a Pull Request

### License

MIT License - see [LICENSE](LICENSE) for details.
