# PGX-Goose - Documentación de Uso

## Descripción General

**PGX-Goose** es una herramienta poderosa que realiza ingeniería inversa en bases de datos PostgreSQL para generar automáticamente código fuente Go, incluyendo structs, interfaces de repositorio, implementaciones, mocks y pruebas unitarias.

## Índice

1. [Instalación](#instalación)
2. [Configuración](#configuración)
3. [Uso Básico](#uso-básico)
4. [Configuraciones Avanzadas](#configuraciones-avanzadas)
5. [Ejemplos Prácticos](#ejemplos-prácticos)
6. [Estructura de Archivos Generados](#estructura-de-archivos-generados)
7. [Personalización](#personalización)
8. [Solución de Problemas](#solución-de-problemas)

## Instalación

### Prerrequisitos
- Go 1.19+ instalado
- Acceso a una base de datos PostgreSQL
- Git (para clonar el repositorio)

### Instalación vía Go
```bash
go install github.com/fsvxavier/pgx-goose@latest
```

### Instalación vía Clone
```bash
git clone https://github.com/fsvxavier/pgx-goose.git
cd pgx-goose
go build -o pgx-goose main.go
```

## Configuración

### Archivo de Configuración

pgx-goose busca automáticamente archivos de configuración en el siguiente orden:
1. `pgx-goose-conf.yaml`
2. `pgx-goose-conf.yml`
3. `pgx-goose-conf.json`

### Configuración Básica (pgx-goose-conf.yaml)

```yaml
# Configuración mínima requerida
dsn: "postgres://user:password@localhost:5432/database?sslmode=disable"
schema: "public"
out: "./generated"
mock_provider: "testify"
with_tests: true
```

### Configuración Completa

```yaml
# Cadena de conexión PostgreSQL
dsn: "postgres://user:password@host:5432/database?sslmode=disable"

# Esquema de base de datos a procesar
schema: "public"

# Configuración de directorios de salida
output_dirs:
  base: "./generated"                    # Directorio base
  models: "./generated/models"           # Entidades/modelos
  interfaces: "./generated/interfaces"   # Interfaces de repositorio
  repositories: "./generated/postgres"   # Implementaciones PostgreSQL
  mocks: "./generated/mocks"             # Mocks para pruebas
  tests: "./generated/tests"             # Pruebas de integración

# Configuraciones de generación
mock_provider: "testify"                 # "testify" o "mock"
with_tests: true                         # Generar pruebas unitarias
template_dir: "./custom_templates"       # Plantillas personalizadas (opcional)

# Filtrado de tablas
tables: []                               # Vacío = todas las tablas
ignore_tables:                          # Tablas a ignorar
  - "migrations"
  - "schema_versions"
```

## Uso Básico

### Comando Básico
```bash
# Usar configuración automática
pgx-goose

# Especificar archivo de configuración
pgx-goose --config pgx-goose-conf.yaml

# Sobrescribir configuraciones vía CLI
pgx-goose --dsn "postgres://..." --schema "public" --out "./generated"
```

### Opciones de Línea de Comandos

| Flag | Descripción | Ejemplo |
|------|-------------|---------|
| `--config` | Archivo de configuración | `--config config.yaml` |
| `--dsn` | Cadena de conexión PostgreSQL | `--dsn "postgres://..."` |
| `--schema` | Esquema de base de datos | `--schema "public"` |
| `--out` | Directorio de salida | `--out "./generated"` |
| `--tables` | Tablas específicas | `--tables "users,products"` |
| `--mock-provider` | Proveedor de mocks | `--mock-provider "testify"` |
| `--template-dir` | Directorio de plantillas | `--template-dir "./templates"` |
| `--verbose` | Registro detallado | `--verbose` |
| `--debug` | Registro de depuración | `--debug` |

## Configuraciones Avanzadas

### Diferentes Entornos

#### Desarrollo
```yaml
dsn: "postgres://dev:devpass@localhost:5432/myapp_dev?sslmode=disable"
schema: "public"
out: "./dev_generated"
mock_provider: "testify"
with_tests: false  # Más rápido durante el desarrollo
tables:
  - "users"
  - "products"
```

#### Producción
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

### Microservicios
```yaml
dsn: "postgres://user:pass@db:5432/microservices?sslmode=disable"
schema: "user_service"  # Esquema específico del servicio
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
  - "product_catalog"  # Tablas de otros servicios
  - "orders"
```

## Ejemplos Prácticos

### Ejemplo 1: Configuración Rápida
```bash
# 1. Crear archivo de configuración básico
cat > pgx-goose-conf.yaml << EOF
dsn: "postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable"
schema: "public"
out: "./generated"
mock_provider: "testify"
with_tests: true
EOF

# 2. Generar código
pgx-goose

# 3. Verificar archivos generados
ls -la generated/
```

### Ejemplo 2: Tablas Específicas
```bash
# Generar solo para tablas específicas
pgx-goose --tables "users,products,orders" --verbose
```

### Ejemplo 3: Esquema Personalizado
```bash
# Trabajar con esquema específico
pgx-goose --schema "billing" --out "./billing_generated"
```

### Ejemplo 4: Plantillas Personalizadas
```bash
# Usar plantillas personalizadas
pgx-goose --template-dir "./my_templates" --mock-provider "mock"
```

## Estructura de Archivos Generados

```
generated/
├── models/
│   ├── user.go              # Struct del modelo User
│   ├── product.go           # Struct del modelo Product
│   └── order.go             # Struct del modelo Order
├── interfaces/
│   ├── user_repository.go   # Interfaz UserRepository
│   ├── product_repository.go
│   └── order_repository.go
├── postgres/
│   ├── user_repository.go   # Implementación PostgreSQL
│   ├── product_repository.go
│   └── order_repository.go
├── mocks/
│   ├── user_repository.go   # Mock UserRepository
│   ├── product_repository.go
│   └── order_repository.go
└── tests/
    ├── user_repository_test.go  # Pruebas de integración
    ├── product_repository_test.go
    └── order_repository_test.go
```

### Ejemplos de Código Generado

#### Modelo (models/user.go)
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

#### Interfaz (interfaces/user_repository.go)
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

## Personalización

### Plantillas Personalizadas

1. **Copiar plantillas por defecto:**
   ```bash
   cp -r templates_custom/base ./my_templates
   ```

2. **Modificar según sea necesario:**
   ```bash
   # Editar plantillas en ./my_templates/
   vim my_templates/model.tmpl
   ```

3. **Usar plantillas personalizadas:**
   ```yaml
   template_dir: "./my_templates"
   ```

### Variables de Entorno

Usar variables de entorno en el archivo de configuración:

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

## Solución de Problemas

### Problemas Comunes

#### 1. Error de Conexión
```
Error: failed to connect to database
```
**Solución:** Verificar DSN, credenciales y conectividad de red.

#### 2. Esquema No Encontrado
```
Error: schema "myschema" does not exist
```
**Solución:** Verificar que el esquema existe en la base de datos.

#### 3. Permisos Insuficientes
```
Error: permission denied for schema
```
**Solución:** Asegurar que el usuario tiene permisos de lectura en el esquema.

#### 4. No Se Encontraron Tablas
```
Warning: no tables found in schema
```
**Solución:** Verificar filtros de tablas y que existan tablas en el esquema.

### Depuración

```bash
# Modo verbose para más información
pgx-goose --verbose

# Modo debug para información detallada
pgx-goose --debug
```

### Registros

Los registros se muestran en la consola con marcas de tiempo:

```
time="2025-07-03T21:53:38-03:00" level=info msg="Starting pgx-goose code generation"
time="2025-07-03T21:53:38-03:00" level=info msg="Found configuration file: pgx-goose-conf.yaml"
time="2025-07-03T21:53:38-03:00" level=info msg="Loading configuration from pgx-goose-conf.yaml"
time="2025-07-03T21:53:38-03:00" level=info msg="Using database schema: 'public'"
```

## Integración con Proyectos

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

## Conclusión

pgx-goose simplifica significativamente el desarrollo de aplicaciones Go con PostgreSQL al automatizar la generación de código boilerplate y garantizar consistencia entre el esquema de la base de datos y el código de la aplicación.

Para más ejemplos, consulte la carpeta `examples/` en el repositorio del proyecto.
