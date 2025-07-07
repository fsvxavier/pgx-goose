# PGX-Goose Configuration Examples

This folder contains different example configuration files for PGX-Goose, demonstrating various approaches and usage scenarios.

> 🇺🇸 **English version (current)** | 🇧🇷 **[Versão em português disponível](README-pt-br.md)** | 🇪🇸 **[Versión en español disponible](README-es.md)**

## Available Files

### Basic Configurations
- **`pgx-goose-conf_basic.yaml`** - Simple and direct configuration to get started quickly
- **`pgx-goose-conf_basic.json`** - Same basic configuration in JSON format

### Advanced Configurations
- **`pgx-goose-conf_advanced.yaml`** - Complete configuration with separate directories and all options
- **`pgx-goose-conf_separate_dirs.yaml`** - Focus on organization with separate directories by type

### Advanced Feature Configurations
- **`pgx-goose-conf_advanced_optimized.yaml`** - Complete configuration with all advanced features enabled:
  - Parallel generation for improved performance
  - Template optimization and caching
  - Incremental generation for faster updates
  - Cross-schema support
  - Migration generation
  - go:generate integration

### Environment-Specific Configurations
- **`pgx-goose-conf_development.yaml`** - Optimized for local development
- **`pgx-goose-conf_production.yaml`** - Robust configuration for production
- **`pgx-goose-conf_testing.yaml`** - For automated testing and CI/CD

### Architecture-Specific Configurations
- **`pgx-goose-conf_microservice.yaml`** - For microservice projects
- **`pgx-goose-conf_custom_schema.yaml`** - For working with specific schemas

### Filtering Configurations
- **`pgx-goose-conf_ignore_tables.yaml`** - Example of how to ignore specific tables

## How to Use

1. **Copy** the example file that best suits your project
2. **Rename** to `pgx-goose-conf.yaml` or `pgx-goose-conf.json`
3. **Edit** the specific configurations for your project:
   - Database DSN
   - Schema
   - Output directories
   - Specific tables or tables to ignore

## Usage Examples

### Using with specific configuration file:
```bash
pgx-goose --config examples/pgx-goose-conf_basic.yaml
```

### Using with automatic search (rename the file):
```bash
cp examples/pgx-goose-conf_basic.yaml pgx-goose-conf.yaml
pgx-goose
```

## Advanced Features Examples

### Performance Optimization
For large databases with many tables, use the optimized configuration:
```bash
pgx-goose --config examples/pgx-goose-conf_advanced_optimized.yaml
```

This configuration includes:
- **Parallel processing** with 8 workers
- **Template caching** for faster regeneration
- **Incremental generation** to only update changed files
- **Cross-schema support** for multi-schema applications

### Incremental Development Workflow
```bash
# First run - generates all files
pgx-goose --config examples/pgx-goose-conf_advanced_optimized.yaml

# Subsequent runs - only regenerates changed files
pgx-goose --incremental

# Force complete regeneration when needed
pgx-goose --force
```

### Multi-Schema Projects
```bash
# Generate code for multiple schemas
pgx-goose --config examples/pgx-goose-conf_advanced_optimized.yaml --schemas "public,auth,audit"
```

### CI/CD Integration
For automated builds, use go:generate integration:
```bash
# Add to your Go files:
//go:generate pgx-goose --config pgx-goose-conf.yaml

# Then run:
go generate ./...
```

## Configuration File Structure

### Main Fields:
- **`dsn`** - PostgreSQL connection string
- **`schema`** - Database schema to process (default: "public")
- **`out`** - Simple output directory (legacy)
- **`output_dirs`** - Detailed directory configuration
- **`mock_provider`** - Mock provider ("testify" or "mock")
- **`with_tests`** - Whether to generate tests (true/false)
- **`template_dir`** - Custom templates directory (optional)
- **`tables`** - List of specific tables (empty = all)
- **`ignore_tables`** - List of tables to ignore

### Directory Configuration (output_dirs):
- **`base`** - Base directory
- **`models`** - Entities/models
- **`interfaces`** - Repository interfaces
- **`repositories`** - PostgreSQL implementations
- **`mocks`** - Test mocks
- **`tests`** - Integration tests

## Tips

1. **Development Environment**: Use simpler and faster configurations
2. **Production**: Use all validations and tests
3. **Microservices**: Focus on specific schemas
4. **CI/CD**: Use configurations optimized for automated testing
5. **Clean Architecture**: Organize directories according to your project structure

## Environment Variables

You can use environment variables in the DSN:
```yaml
dsn: "postgres://user:${DB_PASSWORD}@${DB_HOST}:5432/mydb"
```
