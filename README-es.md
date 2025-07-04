# pgx-goose

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**pgx-goose** es una herramienta de ingeniería inversa de PostgreSQL que genera automáticamente código Go idiomático incluyendo structs, interfaces de repositorio, implementaciones, mocks y pruebas unitarias. Soporta múltiples esquemas para arquitecturas empresariales complejas.

> 🇪🇸 **Versión en español (actual)** | 🇺🇸 **[English version available](README-en.md)** | 🇧🇷 **[Versão em português disponível](README.md)**

## 📋 Tabla de Contenidos

- [🚀 Características](#-características)
- [📦 Instalación](#-instalación)
- [⚡ Inicio Rápido](#-inicio-rápido)
- [⚙️ Configuración](#️-configuración)
- [📁 Estructura Generada](#-estructura-generada)
- [🎨 Plantillas](#-plantillas)
- [🔧 Referencia CLI](#-referencia-cli)
- [💡 Ejemplos de Uso](#-ejemplos-de-uso)
- [🤝 Contribuir](#-contribuir)
- [ Documentación](#-documentación)

## 🚀 Características

- **🔍 Análisis Completo**: Introspección de esquemas PostgreSQL (tablas, columnas, tipos, PKs, índices, relaciones)
- **🏢 Multi-Esquema**: Soporte para esquemas personalizados para arquitecturas empresariales
- **🤖 Generación Automática**: Crea structs, interfaces, implementaciones, mocks y pruebas
- **📂 Directorios Flexibles**: Configuración de directorios de salida personalizable
- **🎨 Plantillas Personalizables**: Plantillas Go personalizadas + plantillas PostgreSQL optimizadas (incluyendo variaciones simples)
- **🧪 Proveedores de Mock**: Soporte para `testify/mock`, `mock` y `gomock`
- **🎯 Arquitectura Limpia**: Código siguiendo principios de Clean Architecture y SOLID
- **⚡ Operaciones Avanzadas**: Transacciones, operaciones por lotes y borrado suave
- **🔧 CLI Robusto**: Interfaz de línea de comandos completa con validación y logging configurable
- **📝 Configuración Flexible**: Soporte YAML/JSON con precedencia jerárquica

## 📦 Instalación

### Vía go install (Recomendado)

```bash
go install github.com/fsvxavier/pgx-goose@latest
```

### Compilación local

```bash
git clone https://github.com/fsvxavier/isis-golang-lib.git
cd pgx-goose
go build -o pgx-goose .
./pgx-goose --help
```

## 📚 Documentación

Documentación completa disponible en múltiples idiomas:

- 🇪🇸 **[Español](docs/usage-es.md)** - Documentación completa en español
- 🇺🇸 **[English](docs/usage-en.md)** - Documentación completa en inglés  
- 🇧🇷 **[Português (Brasil)](docs/usage-pt-br.md)** - Documentación completa en portugués brasileño
- 📋 **[Referencia Rápida](docs/quick-reference.md)** - Referencia rápida para comandos y configuraciones

### Lo que está cubierto en la documentación:
- Instalación detallada y prerrequisitos
- Configuración completa (YAML/JSON)
- Uso básico y avanzado
- Ejemplos prácticos para diferentes escenarios
- Estructura de archivos generados
- Personalización de plantillas
- Solución de problemas y resolución de errores
- Integración del proyecto (Makefile, CI/CD)

### Ejemplos de Configuración
Consulte la carpeta [examples/](examples/) para:
- Configuraciones básicas y avanzadas
- Configuraciones específicas de entorno (dev, prod, testing)
- Configuraciones de microservicios
- Ejemplos de filtrado de tablas

## ⚡ Inicio Rápido

### 1. Comando Simple
```bash
# Generar código para todas las tablas
pgx-goose --dsn "postgres://user:pass@localhost:5432/mydb"
```

### 2. Con Configuración YAML
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

### 3. Comandos Comunes
```bash
# Tablas específicas
pgx-goose --dsn "..." --tables "users,orders,products"

# Esquema personalizado
pgx-goose --dsn "..." --schema "inventory" --out "./inventory-gen"

# Plantillas PostgreSQL optimizadas
pgx-goose --config pgx-goose-conf.yaml --template-dir "./templates_postgresql"
```

## ⚙️ Configuración

### Archivo de Configuración

#### pgx-goose-conf.yaml (Recomendado)
```yaml
# Conexión
dsn: "postgres://user:pass@localhost:5432/db?sslmode=disable"
schema: "public"  # Esquema personalizado (predeterminado: "public")

# Directorios de salida
output_dirs:
  base: "./generated"                       # Directorio base (predeterminado: ./pgx-goose)
  models: "./internal/domain/entities"      # Structs
  interfaces: "./internal/ports"            # Interfaces
  repositories: "./internal/adapters/db"    # Implementaciones
  mocks: "./tests/mocks"                    # Mocks
  tests: "./tests/integration"              # Pruebas

# Filtros de tabla
tables: []                                  # [] = todas, o ["users", "orders"] 
ignore_tables:                             # Tablas a ignorar
  - "schema_migrations"       # Migraciones Rails/Laravel
  - "ar_internal_metadata"    # Metadatos Rails
  - "goose_db_version"        # Migraciones Goose
  - "migrations"              # Migraciones genéricas
  - "audit_logs"              # Tablas de auditoría/log
  - "sessions"                # Datos de sesión

# Opciones de generación
template_dir: "./templates_postgresql"  # Plantillas optimizadas
mock_provider: "testify"                # "testify", "mock", o "gomock"  
with_tests: true                        # Generar pruebas (predeterminado: true)
```

#### pgx-goose-conf.json (Alternativo)
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

### Opciones de Configuración Detalladas

| Campo | Tipo | Predeterminado | Descripción |
|-------|------|---------|-------------|
| `dsn` | string | **requerido** | Cadena de conexión PostgreSQL |
| `schema` | string | `"public"` | Esquema de base de datos a introspectar |
| `output_dirs.base` | string | `"./pgx-goose"` | Directorio de salida base |
| `output_dirs.models` | string | `"{base}/models"` | Directorio para structs |
| `output_dirs.interfaces` | string | `"{base}/repository/interfaces"` | Directorio para interfaces |
| `output_dirs.repositories` | string | `"{base}/repository/postgres"` | Directorio para implementaciones |
| `output_dirs.mocks` | string | `"{base}/mocks"` | Directorio para mocks |
| `output_dirs.tests` | string | `"{base}/tests"` | Directorio para pruebas |
| `tables` | []string | `[]` (todas) | Lista de tablas específicas |
| `ignore_tables` | []string | `[]` | Lista de tablas a ignorar |
| `template_dir` | string | `""` (incluidas) | Directorio de plantillas personalizadas |
| `mock_provider` | string | `"testify"` | Proveedor de mock: `testify`, `mock`, `gomock` |
| `with_tests` | bool | `true` | Generar archivos de prueba |

### Validación y Reglas

1. **DSN Requerido**: El campo `dsn` siempre es requerido
2. **Conflictos de tabla**: No se permite especificar la misma tabla en `tables` e `ignore_tables`
3. **Proveedores de mock válidos**: Solo se aceptan `testify`, `mock` y `gomock`
4. **Directorios**: Si no se especifica, usar predeterminados relativos a `base`

### Precedencia de Configuración

La configuración sigue una jerarquía de precedencia (mayor a menor):

1. **Flags CLI** (mayor precedencia)
2. **Archivo de configuración** (`--config`)
3. **Valores predeterminados** (menor precedencia)

```bash
# CLI anula cualquier valor del archivo de configuración
pgx-goose --config pgx-goose-conf.yaml --schema "billing" --mock-provider "gomock"
```

### Filtrado de Tablas

#### Modo Inclusivo (Tablas Específicas)
```yaml
tables: ["users", "orders", "products"]  # Solo estas tablas
ignore_tables: []                        # Lista de ignorar debe estar vacía
```

#### Modo Exclusivo (Todas Excepto...)
```yaml
tables: []  # Lista vacía = todas las tablas
ignore_tables: 
  - "schema_migrations"      # Rails/Laravel
  - "ar_internal_metadata"   # ActiveRecord
  - "goose_db_version"       # Migraciones Goose
  - "audit_logs"             # Logs de auditoría
  - "sessions"               # Sesiones de usuario
```

#### Validación de Conflictos
```yaml
# ❌ ERROR: Conflicto detectado - tabla en ambas listas
tables: ["users", "orders"]
ignore_tables: ["users"]  # users aparece en ambas listas

# ✅ OK: Sin conflictos
tables: ["users", "orders"] 
ignore_tables: []
```

### Reglas de Validación

El sistema aplica las siguientes validaciones antes de la ejecución:

| Validación | Descripción | Error |
|------------|-------------|-------|
| **DSN Requerido** | El campo `dsn` debe estar presente | `DSN is required` |
| **Proveedor de mock válido** | Debe ser `testify`, `mock` o `gomock` | `invalid mock provider` |
| **Conflictos de tabla** | La tabla no puede estar en `tables` Y `ignore_tables` | `conflicting table configuration` |
| **Archivo de configuración** | Si se especifica, debe existir y ser válido | `failed to read config file` |
| **Formato de configuración** | Debe ser `.yaml`, `.yml` o `.json` | `unsupported config file format` |

### Configuración de Logging

```bash
# Logging predeterminado (solo advertencias/errores)
pgx-goose --config pgx-goose-conf.yaml

# Verbose (info + advertencias + errores)
pgx-goose --config pgx-goose-conf.yaml --verbose

# Debug (todo)
pgx-goose --config pgx-goose-conf.yaml --debug
```

## 📁 Estructura Generada

### Estructura Predeterminada
```
generated/
├── models/                 # Structs de entidad
│   ├── user.go
│   └── product.go
├── repository/
│   ├── interfaces/         # Interfaces de repositorio
│   │   ├── user_repository.go
│   │   └── product_repository.go
│   └── postgres/           # Implementaciones PostgreSQL
│       ├── user_repository.go
│       └── product_repository.go
├── mocks/                  # Mocks de prueba
│   ├── mock_user_repository.go
│   └── mock_product_repository.go
└── tests/                  # Pruebas unitarias/integración
    ├── user_repository_test.go
    └── product_repository_test.go
```

### Estructura de Directorio Personalizada
```
internal/
├── domain/entities/        # Modelos
├── ports/                  # Interfaces  
└── adapters/postgres/      # Implementaciones
tests/
├── mocks/                  # Mocks
└── integration/            # Pruebas
```

### Tipos de Archivos Generados

| Tipo | Descripción | Contenido |
|------|-------------|---------|
| **Modelos** | Structs de entidad | Tags JSON/DB, validación, métodos de utilidad |
| **Interfaces** | Contratos de repositorio | CRUD, transacciones, operaciones por lotes |
| **Repositorios** | Implementaciones PostgreSQL | Pools de conexión, declaraciones preparadas |
| **Mocks** | Testify/GoMock | Métodos de expectativa, aserciones |
| **Pruebas** | Pruebas de integración | Setup/teardown, benchmarks, testcontainers |

## 🎨 Plantillas

### Plantillas Disponibles

#### 1. Plantillas Predeterminadas (`./templates/`)
- Plantillas genéricas para cualquier proyecto Go
- Compatibilidad básica con pgx

#### 2. Plantillas PostgreSQL (`./templates_postgresql/`)
- **Recomendadas** - Optimizadas para `isis-golang-lib`
- Soporte para transacciones y operaciones por lotes
- Métodos avanzados de struct

#### 3. Variaciones de Plantilla

Cada conjunto de plantillas tiene dos variaciones:

| Plantilla | Predeterminada | Simple (`*_simple.tmpl`) |
|----------|---------|-------------------------|
| **Modelo** | Struct completo con métodos de utilidad | Solo struct básico |
| **Repositorio** | Interface/implementación completa | Solo operaciones CRUD básicas |
| **Mock** | Mock completo con todos los métodos | Mock simplificado |
| **Prueba** | Pruebas exhaustivas con benchmarks | Pruebas unitarias básicas |

### Usar Plantillas PostgreSQL
```bash
pgx-goose --template-dir "./templates_postgresql" --config pgx-goose-conf.yaml
```

### Plantillas Personalizadas

Crear un directorio con:
```
my_templates/
├── model.tmpl                  # Structs
├── repository_interface.tmpl   # Interfaces
├── repository_postgres.tmpl    # Implementaciones
├── mock_testify.tmpl          # Mocks Testify
├── mock_gomock.tmpl           # Mocks GoMock
└── test.tmpl                  # Pruebas
```

**Ejemplo model.tmpl:**
```go
package {{.Package}}

import "time"

// {{.StructName}} representa {{.Table.Comment}}
type {{.StructName}} struct {
{{- range .Table.Columns}}
    {{toPascalCase .Name}} {{.GoType}} `json:"{{.Name}}" db:"{{.Name}}"`
{{- end}}
}

func (e *{{.StructName}}) TableName() string {
    return "{{.Table.Name}}"
}
```

## 🔧 Referencia CLI

### Flags Principales

| Flag | Descripción | Valores | Predeterminado |
|------|-------------|--------|---------|
| `--dsn` | Cadena de conexión PostgreSQL | `postgres://user:pass@host:port/db` | **requerido** |
| `--schema` | Esquema de base de datos | `public`, `inventory`, `billing` | `public` |
| `--config` | Archivo de configuración | `pgx-goose-conf.yaml`, `pgx-goose-conf.json` | - |
| `--out` | Directorio de salida | `./generated` | `./pgx-goose` |
| `--tables` | Tablas específicas (CSV) | `users,orders,products` | todas |
| `--template-dir` | Directorio de plantillas | `./templates_postgresql` | incluidas |
| `--mock-provider` | Proveedor de mock | `testify`, `mock`, `gomock` | `testify` |
| `--with-tests` | Generar pruebas | `true`, `false` | `true` |

### Flags de Directorio Específico

| Flag | Directorio | Ejemplo |
|------|-----------|---------|
| `--models-dir` | Modelos/structs | `./internal/domain/entities` |
| `--interfaces-dir` | Interfaces | `./internal/ports` |
| `--repos-dir` | Implementaciones | `./internal/adapters/postgres` |
| `--mocks-dir` | Mocks | `./tests/mocks` |
| `--tests-dir` | Pruebas | `./tests/integration` |

### Flags de Configuración y Logging

| Flag | Descripción | Uso |
|------|-------------|-------|
| `--json` | Usar formato JSON para configuración | Para preferir .json sobre .yaml |
| `--yaml` | Usar formato YAML para configuración | Predeterminado, explícito |
| `--verbose` | Logging verbose (nivel INFO) | Depuración de ejecución |
| `--debug` | Logging debug (nivel DEBUG) | Depuración completa |

### Ejemplos de Comandos

```bash
# Básico
pgx-goose --dsn "postgres://user:pass@localhost:5432/db"

# Esquema personalizado + tablas específicas
pgx-goose --dsn "..." --schema "billing" --tables "invoices,payments"

# Configuración completa con logging
pgx-goose --config pgx-goose-conf.yaml --template-dir "./templates_postgresql" --verbose

# Ignorar tablas específicas
pgx-goose --dsn "..." --ignore-tables "migrations,logs,sessions"

# Organización modular con directorios personalizados
pgx-goose --dsn "..." --tables "users" \
  --models-dir "./modules/user/entity" \
  --interfaces-dir "./modules/user/repository"

# Multi-esquema empresarial
pgx-goose --schema "inventory" --out "./modules/inventory/generated"
pgx-goose --schema "billing" --out "./modules/billing/generated"

# Proveedor de mock personalizado
pgx-goose --config pgx-goose-conf.yaml --mock-provider "gomock" --debug
```

## 💡 Ejemplos de Uso

### 1. Proyecto de E-commerce Simple

**Esquema SQL:**
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

**Configuración:**
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

**Generar código:**
```bash
pgx-goose --config pgx-goose-conf.yaml
```

### 2. Arquitectura Multi-Esquema Empresarial

```bash
# Esquema de usuarios
pgx-goose --schema "users" --out "./modules/users/generated"

# Esquema de inventario
pgx-goose --schema "inventory" --out "./modules/inventory/generated"

# Esquema de facturación
pgx-goose --schema "billing" --out "./modules/billing/generated"
```

### 3. Código Generado - Ejemplo Usuario

**Modelo Generado (`models/user.go`):**
```go
type User struct {
    ID        int64     `json:"id" db:"id"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    Name      string    `json:"name" db:"name" validate:"required"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (u *User) TableName() string { return "users" }
func (u *User) Validate() error { /* validación */ }
func (u *User) Clone() *User { /* clonación segura */ }
```

**Interface Generada (`interfaces/user_repository.go`):**
```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id int64) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id int64) error
    
    // Transacciones
    CreateTx(ctx context.Context, tx common.ITransaction, user *models.User) error
    
    // Operaciones por lotes
    CreateBatch(ctx context.Context, users []*models.User) error
    
    // Búsquedas específicas
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}
```

### 4. Usando Código Generado

```go
package main

import (
    "context"
    "your-project/internal/domain"
    "your-project/internal/adapters/postgres"
    "github.com/fsvxavier/isis-golang-lib/db/postgresql"
)

func main() {
    // Configurar pool PostgreSQL
    pool, _ := postgresql.NewPool(postgresql.Config{
        Host:     "localhost",
        Database: "ecommerce",
        Username: "admin",
        Password: "password",
    })
    defer pool.Close()
    
    // Usar repositorio generado
    userRepo := postgres.NewUserRepository(pool)
    
    // Crear usuario
    user := &domain.User{
        Email: "john@example.com",
        Name:  "John Doe",
    }
    
    err := userRepo.Create(context.Background(), user)
    if err != nil {
        panic(err)
    }
    
    // Buscar por email
    found, _ := userRepo.FindByEmail(context.Background(), "john@example.com")
    fmt.Printf("Usuario encontrado: %+v\n", found)
}
```

### 5. Pruebas con Mocks

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

## 🤝 Contribuir

### Cómo Contribuir

1. **Fork** el proyecto
2. **Crear una rama** (`git checkout -b feature/NuevaCaracteristica`)
3. **Commit** tus cambios (`git commit -m 'Add: nueva característica'`)
4. **Push** a la rama (`git push origin feature/NuevaCaracteristica`)
5. **Abrir un Pull Request**

### Desarrollo Local

```bash
# Clonar y configurar
git clone https://github.com/fsvxavier/isis-golang-lib.git
cd pgx-goose
go mod download

# Pruebas
go test ./...

# Compilar
go build -o pgx-goose .
./pgx-goose --help
```

### Estructura del Proyecto

```
pgx-goose/
├── cmd/                    # Comandos CLI (Cobra)
├── internal/
│   ├── config/            # Configuración
│   ├── generator/         # Generación de código
│   └── introspector/      # Introspección PostgreSQL
├── templates/             # Plantillas predeterminadas
├── templates_postgresql/  # Plantillas optimizadas
├── examples/              # Ejemplos de configuración
└── docs/                  # Documentación adicional
```

### Pautas

- **Pruebas**: Toda nueva característica debe tener pruebas
- **Documentación**: Actualizar README.md para nuevas características
- **Plantillas**: Mantener compatibilidad con plantillas existentes
- **Logs**: Usar slog para logging estructurado

---

## 📄 Licencia

Licenciado bajo la [Licencia MIT](LICENSE).

## 🙏 Reconocimientos

- [pgx](https://github.com/jackc/pgx) - Driver PostgreSQL de alto rendimiento
- [Cobra](https://github.com/spf13/cobra) - Framework CLI
- [testify](https://github.com/stretchr/testify) - Framework de pruebas
- [testcontainers](https://github.com/testcontainers/testcontainers-go) - Pruebas de integración

## 📞 Soporte

- **Issues**: [GitHub Issues](https://github.com/fsvxavier/isis-golang-lib/issues)
- **Discusiones**: [GitHub Discussions](https://github.com/fsvxavier/isis-golang-lib/discussions)

---

**pgx-goose** - ¡Transformando tu PostgreSQL en código Go idiomático! 🚀
