# Registro de Cambios

Todos los cambios notables de este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
y este proyecto se adhiere al [Versionado Semántico](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-03

### 🎯 Funcionalidades Principales Implementadas

#### ✅ Directorios de Salida Configurables (NUEVA FUNCIONALIDAD)
- Agregada configuración `output_dirs` para directorios separados
- Soporte para flags CLI específicos (`--models-dir`, `--interfaces-dir`, etc.)
- Compatibilidad hacia atrás con configuración `OutputDir` legacy
- Precedencia: CLI flags > archivo de configuración > valores por defecto
- Soporte para 5 arquitecturas diferentes (Hexagonal, Clean, DDD, Modular, Monorepo)

#### ✅ Templates PostgreSQL Optimizados (NUEVA FUNCIONALIDAD)
- Templates especializados en carpeta `templates_postgresql/`
- Integración con proveedor `db/postgresql` de nexs-lib
- Soporte para operaciones transaccionales y por lotes
- Métodos avanzados en entidades (TableName, Clone, Validate, etc.)
- Eliminación lógica cuando aplica

#### ✅ Documentación Unificada (NUEVA FUNCIONALIDAD)
- README.md completo con ~1,315 líneas
- 17 secciones principales con 15+ ejemplos completos
- Integración de 6 archivos de documentación
- Índice detallado para navegación
- Casos de uso por arquitectura

### Agregado

#### Funcionalidades Principales
- ✅ Herramienta CLI completa basada en Cobra
- ✅ Introspección completa de esquemas PostgreSQL
- ✅ Generación automática de structs Go a partir de tablas
- ✅ Generación de interfaces de repositorios con CRUD completo
- ✅ Implementaciones PostgreSQL usando pgx/v5
- ✅ Soporte para dos proveedores de mock: testify y gomock
- ✅ Generación automática de pruebas unitarias
- ✅ Sistema de templates personalizable usando Go Templates

#### Configuración y CLI
- ✅ Soporte para archivos de configuración YAML y JSON
- ✅ Flags de línea de comandos completos
- ✅ Sistema de logging configurable (debug, verbose, info, warn, error)
- ✅ Validación de configuración robusta

#### Soporte de Base de Datos
- ✅ Introspección completa de tablas, columnas, tipos
- ✅ Soporte para llaves primarias, índices y llaves foráneas
- ✅ Mapeo automático de tipos PostgreSQL → Go
- ✅ Soporte para tipos nullable con punteros
- ✅ Comentarios de tablas y columnas preservados

#### Generación de Código
- ✅ Templates embebidos para todos los tipos de archivo
- ✅ Estructura de proyecto organizada e idiomática
- ✅ Soporte para templates personalizados vía directorio customizado
- ✅ Generación de código limpio siguiendo convenciones Go
- ✅ Soporte para relaciones entre tablas

#### Pruebas y Calidad
- ✅ Pruebas unitarias exhaustivas
- ✅ Mocks automáticos para todas las interfaces
- ✅ Pruebas generadas con escenarios de éxito y error
- ✅ Integración con testify/assert y testify/mock
- ✅ Soporte para gomock para proyectos que lo prefieran

### Características Técnicas

#### Arquitectura
- Clean Architecture con separación de responsabilidades
- Diseño orientado a interfaces
- Inyección de dependencias
- Estructura modular y extensible

#### Tipos Soportados
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

#### Estructura del Proyecto Generado
```
output_dir/
├── models/                     # Structs de tablas
├── repository/
│   ├── interfaces/             # Interfaces de repositorios
│   └── postgres/               # Implementaciones PostgreSQL
├── mocks/                      # Mocks para pruebas
└── tests/                      # Pruebas unitarias
```

#### Flags CLI Disponibles
- `--dsn`: Cadena de conexión PostgreSQL
- `--out`: Directorio de salida
- `--tables`: Lista de tablas específicas
- `--config`: Archivo de configuración
- `--template-dir`: Templates personalizados
- `--mock-provider`: Proveedor de mocks (testify/mock)
- `--with-tests`: Generación de pruebas
- `--verbose`: Logging verboso
- `--debug`: Logging de debug

### Archivos del Proyecto

#### Documentación
- ✅ README.md completo con ejemplos
- ✅ EXAMPLES.md con casos de uso detallados
- ✅ Templates de configuración (YAML/JSON)
- ✅ Scripts de demostración (Bash/PowerShell)

#### Build y Desarrollo
- ✅ Makefile con targets útiles
- ✅ go.mod con todas las dependencias
- ✅ .gitignore apropiado
- ✅ Licencia MIT

#### Scripts y Herramientas
- ✅ demo.sh para Linux/macOS
- ✅ demo.ps1 para Windows
- ✅ Esquema SQL de ejemplo
- ✅ Configuraciones de ejemplo

### Dependencias

#### Runtime
- `github.com/jackc/pgx/v5` - Driver PostgreSQL
- `github.com/spf13/cobra` - Framework CLI
- `log/slog` - Logging estructurado (nativo Go 1.21+)
- `gopkg.in/yaml.v3` - Parser YAML

#### Pruebas y Mocking
- `github.com/stretchr/testify` - Framework de pruebas
- `go.uber.org/mock` - Gomock para mocks
- `github.com/google/uuid` - Soporte UUID
- `github.com/shopspring/decimal` - Tipos decimales

### Notas de Uso

#### Instalación
```bash
go install github.com/fsvxavier/pgx-goose@latest
```

#### Uso Básico
```bash
pgx-goose --dsn "postgres://user:pass@localhost:5432/db" --out ./generated
```

#### Con Configuración
```bash
pgx-goose --config pgx-goose-conf.yaml --verbose
```

### Reconocimientos

Este proyecto fue inspirado por:
- [xo/dbtpl](https://github.com/xo/dbtpl)
- [go-gorm/gen](https://github.com/go-gorm/gen)
- Principios de Clean Architecture
- Patrones SOLID y DDD

### Contribuir

Para contribuir al proyecto:
1. Haz fork del repositorio
2. Crea una rama para tu característica
3. Implementa con pruebas
4. Ejecuta pruebas y linting
5. Envía un Pull Request

### Licencia

Licencia MIT - ver [LICENSE](LICENSE) para detalles.
