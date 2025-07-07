# Ejemplos de Configuración PGX-Goose

Esta carpeta contiene diferentes archivos de ejemplo de configuración para PGX-Goose, demostrando varios enfoques y escenarios de uso.

> 🇪🇸 **Versión en español (actual)** | 🇧🇷 **[Versão em português disponível](README-pt-br.md)** | 🇺🇸 **[English version available](README.md)**

## Archivos Disponibles

### Configuraciones Básicas
- **`pgx-goose-conf_basic.yaml`** - Configuración simple y directa para empezar rápidamente
- **`pgx-goose-conf_basic.json`** - Misma configuración básica en formato JSON

### Configuraciones Avanzadas
- **`pgx-goose-conf_advanced.yaml`** - Configuración completa con directorios separados y todas las opciones
- **`pgx-goose-conf_separate_dirs.yaml`** - Enfoque en organización con directorios separados por tipo

### Configuraciones de Funcionalidades Avanzadas
- **`pgx-goose-conf_advanced_optimized.yaml`** - Configuración completa con todas las funcionalidades avanzadas habilitadas:
  - Generación paralela para mejor rendimiento
  - Optimización y caché de plantillas
  - Generación incremental para actualizaciones más rápidas
  - Soporte cross-schema
  - Generación de migraciones
  - Integración go:generate

### Configuraciones por Entorno
- **`pgx-goose-conf_development.yaml`** - Optimizada para desarrollo local
- **`pgx-goose-conf_production.yaml`** - Configuración robusta para producción
- **`pgx-goose-conf_testing.yaml`** - Para pruebas automatizadas y CI/CD

### Configuraciones por Arquitectura
- **`pgx-goose-conf_microservice.yaml`** - Para proyectos de microservicios
- **`pgx-goose-conf_custom_schema.yaml`** - Para trabajar con esquemas específicos

### Configuraciones por Filtrado
- **`pgx-goose-conf_ignore_tables.yaml`** - Ejemplo de cómo ignorar tablas específicas

## Cómo Usar

1. **Copie** el archivo de ejemplo que mejor se adapte a su proyecto
2. **Renombre** a `pgx-goose-conf.yaml` o `pgx-goose-conf.json`
3. **Edite** las configuraciones específicas de su proyecto:
   - DSN de la base de datos
   - Esquema
   - Directorios de salida
   - Tablas específicas o tablas a ignorar

## Ejemplos de Uso

### Uso con archivo de configuración específico:
```bash
pgx-goose --config examples/pgx-goose-conf_basic.yaml
```

### Uso con búsqueda automática (renombre el archivo):
```bash
cp examples/pgx-goose-conf_basic.yaml pgx-goose-conf.yaml
pgx-goose
```

## Ejemplos de Funcionalidades Avanzadas

### Optimización de Rendimiento
Para bases de datos grandes con muchas tablas, use la configuración optimizada:
```bash
pgx-goose --config examples/pgx-goose-conf_advanced_optimized.yaml
```

Esta configuración incluye:
- **Procesamiento paralelo** con 8 workers
- **Caché de plantillas** para regeneración más rápida
- **Generación incremental** para actualizar solo archivos cambiados
- **Soporte cross-schema** para aplicaciones multi-esquema

### Flujo de Desarrollo Incremental
```bash
# Primera ejecución - genera todos los archivos
pgx-goose --config examples/pgx-goose-conf_advanced_optimized.yaml

# Ejecuciones posteriores - solo regenera archivos modificados
pgx-goose --incremental

# Forzar regeneración completa cuando sea necesario
pgx-goose --force
```

### Proyectos Multi-Schema
```bash
# Generar código para múltiples esquemas
pgx-goose --config examples/pgx-goose-conf_advanced_optimized.yaml --schemas "public,auth,audit"
```

### Integración CI/CD
Para builds automatizados, use la integración go:generate:
```bash
# Agregue a sus archivos Go:
//go:generate pgx-goose --config pgx-goose-conf.yaml

# Luego ejecute:
go generate ./...
```

## Estructura de Archivos de Configuración

### Campos Principales:
- **`dsn`** - Cadena de conexión PostgreSQL
- **`schema`** - Esquema de base de datos a procesar (predeterminado: "public")
- **`out`** - Directorio de salida simple (legado)
- **`output_dirs`** - Configuración detallada de directorios
- **`mock_provider`** - Proveedor de mocks ("testify" o "mock")
- **`with_tests`** - Si debe generar pruebas (true/false)
- **`template_dir`** - Directorio de plantillas personalizadas (opcional)
- **`tables`** - Lista de tablas específicas (vacío = todas)
- **`ignore_tables`** - Lista de tablas a ignorar

### Configuración de Directorios (output_dirs):
- **`base`** - Directorio base
- **`models`** - Entidades/modelos
- **`interfaces`** - Interfaces de repositorio
- **`repositories`** - Implementaciones PostgreSQL
- **`mocks`** - Mocks para pruebas
- **`tests`** - Pruebas de integración

## Consejos

1. **Entorno de Desarrollo**: Use configuraciones más simples y rápidas
2. **Producción**: Use todas las validaciones y pruebas
3. **Microservicios**: Enfóquese en esquemas específicos
4. **CI/CD**: Use configuraciones optimizadas para pruebas automatizadas
5. **Clean Architecture**: Organice los directorios según la estructura de su proyecto

## Variables de Entorno

Puede usar variables de entorno en el DSN:
```yaml
dsn: "postgres://user:${DB_PASSWORD}@${DB_HOST}:5432/mydb"
```
