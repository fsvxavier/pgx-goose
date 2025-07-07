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

## Funcionalidades Avanzadas

PGX-Goose ofrece varias funcionalidades avanzadas para optimizar la generación de código y mejorar el flujo de trabajo de desarrollo:

### 1. Generación Paralela

**Descripción:** Acelera la generación de código procesando múltiples tablas de forma concurrente.

**Beneficios:**
- Reduce significativamente el tiempo de generación para bases de datos grandes
- Utilización óptima de CPU
- Número de workers configurable

**Configuración:**
```yaml
# Habilitar generación paralela
parallel:
  enabled: true
  workers: 4  # Número de workers concurrentes (predeterminado: núcleos de CPU)
```

**Línea de Comandos:**
```bash
pgx-goose --parallel --workers 8
```

### 2. Optimización de Plantillas y Caché

**Descripción:** Sistema de caché inteligente para plantillas compiladas para mejorar el rendimiento.

**Beneficios:**
- Compilación de plantillas más rápida en ejecuciones posteriores
- Menor uso de memoria
- Tamaño de caché configurable

**Configuración:**
```yaml
template_optimization:
  enabled: true
  cache_size: 100
  precompile: true
```

### 3. Generación Incremental

**Descripción:** Solo regenera archivos que han cambiado, ahorrando tiempo y preservando modificaciones manuales.

**Beneficios:**
- Generación más rápida para proyectos grandes
- Preserva cambios manuales en archivos generados
- Detección inteligente de cambios basada en hash de esquema

**Configuración:**
```yaml
incremental:
  enabled: true
  force: false  # Establecer en true para forzar regeneración completa
```

**Línea de Comandos:**
```bash
pgx-goose --incremental
pgx-goose --force  # Forzar regeneración completa
```

### 4. Soporte Cross-Schema

**Descripción:** Genera código para tablas a través de múltiples esquemas PostgreSQL con detección automática de relaciones.

**Beneficios:**
- Soporte para aplicaciones multi-esquema
- Detección automática de relaciones de claves foráneas entre esquemas
- Generación de código organizada por esquema

**Configuración:**
```yaml
cross_schema:
  enabled: true
  schemas:
    - "public"
    - "auth"
    - "audit"
  relationship_detection: true
```

### 5. Generación de Migraciones

**Descripción:** Genera automáticamente migraciones SQL compatibles con Goose a partir de cambios de esquema.

**Beneficios:**
- Creación automática de migraciones de base de datos
- Soporte para formato de migración Goose
- Detección de cambios y generación de SQL

**Configuración:**
```yaml
migrations:
  enabled: true
  output_dir: "./migrations"
  format: "goose"  # Actualmente soporta "goose"
  naming_pattern: "20060102150405_{{.name}}.sql"
```

**Línea de Comandos:**
```bash
pgx-goose --migrations --migration-dir ./db/migrations
```

### 6. Integración go:generate

**Descripción:** Integración perfecta con la directiva `go:generate` de Go para builds automatizados.

**Beneficios:**
- Generación automática de código durante builds
- Integración con herramientas de desarrollo
- Automatización de tareas VS Code

**Configuración:**
```go
//go:generate pgx-goose --config pgx-goose-conf.yaml
package main
```

**Configuración:**
```yaml
go_generate:
  enabled: true
  create_directive: true
  update_makefile: true
  update_vscode_tasks: true
  update_gitignore: true
```

## Optimización de Rendimiento

### Mejores Prácticas para Bases de Datos Grandes

1. **Habilitar Procesamiento Paralelo:**
   ```yaml
   parallel:
     enabled: true
     workers: 8  # Ajustar según los núcleos de CPU
   ```

2. **Usar Generación Incremental:**
   ```yaml
   incremental:
     enabled: true
   ```

3. **Optimizar Caché de Plantillas:**
   ```yaml
   template_optimization:
     enabled: true
     cache_size: 200
     precompile: true
   ```

4. **Filtrar Tablas Estratégicamente:**
   ```yaml
   ignore_tables:
     - "*_temp"
     - "*_backup"
     - "audit_*"
   ```

### Comparación de Rendimiento

| Funcionalidad | Sin Optimización | Con Todas las Funcionalidades |
|--------------|-----------------|------------------------------|
| 100 tablas | ~45 segundos | ~8 segundos |
| 500 tablas | ~3.5 minutos | ~25 segundos |
| 1000 tablas | ~7 minutos | ~45 segundos |
