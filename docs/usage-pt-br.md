# PGX-Goose - Documentação de Utilização

## Visão Geral

O **PGX-Goose** é uma ferramenta poderosa que realiza engenharia reversa em bancos de dados PostgreSQL para gerar automaticamente código Go, incluindo structs, interfaces de repositório, implementações, mocks e testes unitários.

## Índice

1. [Instalação](#instalação)
2. [Configuração](#configuração)
3. [Uso Básico](#uso-básico)
4. [Configurações Avançadas](#configurações-avançadas)
5. [Exemplos Práticos](#exemplos-práticos)
6. [Estrutura de Arquivos Gerados](#estrutura-de-arquivos-gerados)
7. [Personalização](#personalização)
8. [Troubleshooting](#troubleshooting)

## Instalação

### Pré-requisitos
- Go 1.19+ instalado
- Acesso a um banco PostgreSQL
- Git (para clone do repositório)

### Instalação via Go
```bash
go install github.com/fsvxavier/pgx-goose@latest
```

### Instalação via Clone
```bash
git clone https://github.com/fsvxavier/pgx-goose.git
cd pgx-goose
go build -o pgx-goose main.go
```

## Configuração

### Arquivo de Configuração

O pgx-goose procura automaticamente por arquivos de configuração na seguinte ordem:
1. `pgx-goose-conf.yaml`
2. `pgx-goose-conf.yml`
3. `pgx-goose-conf.json`

### Configuração Básica (pgx-goose-conf.yaml)

```yaml
# Configuração mínima necessária
dsn: "postgres://user:password@localhost:5432/database?sslmode=disable"
schema: "public"
out: "./generated"
mock_provider: "testify"
with_tests: true
```

### Configuração Completa

```yaml
# String de conexão PostgreSQL
dsn: "postgres://user:password@host:5432/database?sslmode=disable"

# Schema do banco a ser processado
schema: "public"

# Configuração de diretórios de saída
output_dirs:
  base: "./generated"                    # Diretório base
  models: "./generated/models"           # Entidades/modelos
  interfaces: "./generated/interfaces"   # Interfaces dos repositórios
  repositories: "./generated/postgres"   # Implementações PostgreSQL
  mocks: "./generated/mocks"             # Mocks para testes
  tests: "./generated/tests"             # Testes de integração

# Configurações de geração
mock_provider: "testify"                 # "testify" ou "mock"
with_tests: true                         # Gerar testes unitários
template_dir: "./custom_templates"       # Templates personalizados (opcional)

# Filtragem de tabelas
tables: []                               # Vazio = todas as tabelas
ignore_tables:                          # Tabelas a ignorar
  - "migrations"
  - "schema_versions"
```

## Uso Básico

### Comando Básico
```bash
# Usar configuração automática
pgx-goose

# Especificar arquivo de configuração
pgx-goose --config pgx-goose-conf.yaml

# Sobrescrever configurações via CLI
pgx-goose --dsn "postgres://..." --schema "public" --out "./generated"
```

### Opções de Linha de Comando

| Flag | Descrição | Exemplo |
|------|-----------|---------|
| `--config` | Arquivo de configuração | `--config config.yaml` |
| `--dsn` | String de conexão PostgreSQL | `--dsn "postgres://..."` |
| `--schema` | Schema do banco | `--schema "public"` |
| `--out` | Diretório de saída | `--out "./generated"` |
| `--tables` | Tabelas específicas | `--tables "users,products"` |
| `--mock-provider` | Provider de mocks | `--mock-provider "testify"` |
| `--template-dir` | Diretório de templates | `--template-dir "./templates"` |
| `--verbose` | Log detalhado | `--verbose` |
| `--debug` | Log de debug | `--debug` |

## Configurações Avançadas

### Diferentes Ambientes

#### Desenvolvimento
```yaml
dsn: "postgres://dev:devpass@localhost:5432/myapp_dev?sslmode=disable"
schema: "public"
out: "./dev_generated"
mock_provider: "testify"
with_tests: false  # Mais rápido durante desenvolvimento
tables:
  - "users"
  - "products"
```

#### Produção
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

### Microserviços
```yaml
dsn: "postgres://user:pass@db:5432/microservices?sslmode=disable"
schema: "user_service"  # Schema específico do serviço
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
  - "product_catalog"  # Tabelas de outros serviços
  - "orders"
```

## Exemplos Práticos

### Exemplo 1: Setup Rápido
```bash
# 1. Criar arquivo de configuração básico
cat > pgx-goose-conf.yaml << EOF
dsn: "postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable"
schema: "public"
out: "./generated"
mock_provider: "testify"
with_tests: true
EOF

# 2. Gerar código
pgx-goose

# 3. Verificar arquivos gerados
ls -la generated/
```

### Exemplo 2: Tabelas Específicas
```bash
# Gerar apenas para tabelas específicas
pgx-goose --tables "users,products,orders" --verbose
```

### Exemplo 3: Schema Customizado
```bash
# Trabalhar com schema específico
pgx-goose --schema "billing" --out "./billing_generated"
```

### Exemplo 4: Templates Personalizados
```bash
# Usar templates customizados
pgx-goose --template-dir "./my_templates" --mock-provider "mock"
```

## Estrutura de Arquivos Gerados

```
generated/
├── models/
│   ├── user.go              # Struct do modelo User
│   ├── product.go           # Struct do modelo Product
│   └── order.go             # Struct do modelo Order
├── interfaces/
│   ├── user_repository.go   # Interface UserRepository
│   ├── product_repository.go
│   └── order_repository.go
├── postgres/
│   ├── user_repository.go   # Implementação PostgreSQL
│   ├── product_repository.go
│   └── order_repository.go
├── mocks/
│   ├── user_repository.go   # Mock UserRepository
│   ├── product_repository.go
│   └── order_repository.go
└── tests/
    ├── user_repository_test.go  # Testes de integração
    ├── product_repository_test.go
    └── order_repository_test.go
```

### Exemplo de Código Gerado

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

## Personalização

### Templates Personalizados

1. **Copiar templates padrão:**
   ```bash
   cp -r templates_custom/base ./my_templates
   ```

2. **Modificar conforme necessário:**
   ```bash
   # Editar templates em ./my_templates/
   vim my_templates/model.tmpl
   ```

3. **Usar templates personalizados:**
   ```yaml
   template_dir: "./my_templates"
   ```

### Variáveis de Ambiente

Usar variáveis de ambiente no arquivo de configuração:

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

### Problemas Comuns

#### 1. Erro de Conexão
```
Error: failed to connect to database
```
**Solução:** Verificar DSN, credenciais e conectividade de rede.

#### 2. Schema não encontrado
```
Error: schema "myschema" does not exist
```
**Solução:** Verificar se o schema existe no banco de dados.

#### 3. Permissões insuficientes
```
Error: permission denied for schema
```
**Solução:** Garantir que o usuário tem permissões de leitura no schema.

#### 4. Tabelas não encontradas
```
Warning: no tables found in schema
```
**Solução:** Verificar filtros de tabelas e se existem tabelas no schema.

### Debug

```bash
# Modo verbose para mais informações
pgx-goose --verbose

# Modo debug para informações detalhadas
pgx-goose --debug
```

### Logs

Os logs são exibidos no console com timestamps:

```
time="2025-07-03T21:53:38-03:00" level=info msg="Starting pgx-goose code generation"
time="2025-07-03T21:53:38-03:00" level=info msg="Found configuration file: pgx-goose-conf.yaml"
time="2025-07-03T21:53:38-03:00" level=info msg="Loading configuration from pgx-goose-conf.yaml"
time="2025-07-03T21:53:38-03:00" level=info msg="Using database schema: 'public'"
```

## Integração com Projetos

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

## Conclusão

O pgx-goose simplifica significativamente o desenvolvimento de aplicações Go com PostgreSQL, automatizando a geração de código boilerplate e garantindo consistência entre o schema do banco e o código da aplicação.

Para mais exemplos, consulte a pasta `examples/` no repositório do projeto.
