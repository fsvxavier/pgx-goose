# pgx-goose
- [⚙️ Configuração](#️-configuração)
- [📁 Estrutura Gerada](#-estrutura-gerada)
- [🎨 Templates](#-templates)
- [🔧 Referência CLI](#-referência-cli)
- [💡 Exemplos de Uso](#-exemplos-de-uso)
- [🤝 Contribuindo](#-contribuindo)
- [📚 Documentação](#-documentação)ersion](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**pgx-goose** é uma ferramenta de engenharia reversa para PostgreSQL que gera automaticamente código Go idiomático incluindo structs, interfaces de repositórios, implementações, mocks e testes unitários. Suporta múltiplos schemas para arquiteturas empresariais complexas.

> 🇺🇸 **[English version available](README.md)** | 🇧🇷 **Versão em português (atual)** | 🇪🇸 **[Versión en español disponible](README-es.md)**

## 📋 Índice

- [🚀 Características](#-características)
- [📦 Instalação](#-instalação)
- [⚡ Quick Start](#-quick-start)
- [⚙️ Configuração](#️-configuração)
- [📁 Estrutura Gerada](#-estrutura-gerada)
- [🎨 Templates](#-templates)
- [🔧 Referência da CLI](#-referência-da-cli)
- [💡 Exemplos de Uso](#-exemplos-de-uso)
- [🤝 Contribuição](#-contribuição)
- [ Documentação](#-documentação)

## 🚀 Características

- **🔍 Análise Completa**: Introspecciona schemas PostgreSQL (tabelas, colunas, tipos, PKs, índices, relacionamentos)
- **🏢 Multi-Schema**: Suporte a schemas personalizados para arquiteturas empresariais
- **🤖 Geração Automática**: Cria structs, interfaces, implementações, mocks e testes
- **📂 Diretórios Flexíveis**: Configuração customizável de diretórios de saída
- **🎨 Templates Customizáveis**: Templates Go personalizados + templates PostgreSQL otimizados (incluindo variações simples)
- **🧪 Mock Providers**: Suporte a `testify/mock`, `mock` e `gomock`
- **🎯 Arquitetura Limpa**: Código seguindo Clean Architecture e SOLID
- **⚡ Operações Avançadas**: Transações, operações em lote e soft delete
- **🔧 CLI Robusta**: Interface de linha de comando completa com validação e logging configurável
- **📝 Configuração Flexível**: Suporte a YAML/JSON com precedência hierárquica

## 📦 Instalação

### Via go install (Recomendado)

```bash
go install github.com/fsvxavier/pgx-goose@latest
```

### Build local

```bash
git clone https://github.com/fsvxavier/isis-golang-lib.git
cd pgx-goose
go build -o pgx-goose .
./pgx-goose --help
```

## 📚 Documentação

Documentação completa disponível em múltiplos idiomas:

- 🇧🇷 **[Português (Brasil)](docs/usage-pt-br.md)** - Documentação completa em português brasileiro
- 🇺🇸 **[English](docs/usage-en.md)** - Complete documentation in English  
- 🇪🇸 **[Español](docs/usage-es.md)** - Documentación completa en español
- 📋 **[Quick Reference](docs/quick-reference.md)** - Referência rápida de comandos e configurações

### O que está coberto na documentação:
- Instalação detalhada e pré-requisitos
- Configuração completa (YAML/JSON)
- Uso básico e avançado
- Exemplos práticos para diferentes cenários
- Estrutura de arquivos gerados
- Personalização com templates
- Troubleshooting e solução de problemas
- Integração com projetos (Makefile, CI/CD)

### Exemplos de Configuração
Veja a pasta [examples/](examples/) para:
- Configurações básicas e avançadas
- Setups específicos por ambiente (dev, prod, testing)
- Configurações para microserviços
- Exemplos de filtragem de tabelas

## ⚡ Quick Start

### 1. Comando Simples
```bash
# Gerar código para todas as tabelas
pgx-goose --dsn "postgres://user:pass@localhost:5432/mydb"
```

### 2. Com Configuração YAML
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

### 3. Comandos Comuns
```bash
# Tabelas específicas
pgx-goose --dsn "..." --tables "users,orders,products"

# Schema customizado
pgx-goose --dsn "..." --schema "inventory" --out "./inventory-gen"

# Templates PostgreSQL otimizados
pgx-goose --config pgx-goose-conf.yaml --template-dir "./templates_postgresql"
```

## ⚙️ Configuração

### Arquivo de Configuração

#### pgx-goose-conf.yaml (Recomendado)
```yaml
# Conexão
dsn: "postgres://user:pass@localhost:5432/db?sslmode=disable"
schema: "public"  # Schema customizado (padrão: "public")

# Diretórios de saída
output_dirs:
  base: "./generated"                       # Diretório base (padrão: ./pgx-goose)
  models: "./internal/domain/entities"      # Structs
  interfaces: "./internal/ports"            # Interfaces
  repositories: "./internal/adapters/db"    # Implementações
  mocks: "./tests/mocks"                    # Mocks
  tests: "./tests/integration"              # Testes

# Filtros de tabelas
tables: []                                  # [] = todas, ou ["users", "orders"] 
ignore_tables: ["migrations", "logs"]      # Tabelas para ignorar

# Configuração de templates e mocks
template_dir: "./templates_postgresql"      # Templates a usar
mock_provider: "testify"                    # "testify", "mock", ou "gomock"  
with_tests: true                           # Gerar testes (padrão: true)
  tests: "./tests/integration"              # Testes

# Filtros de tabelas
tables: []                    # [] = todas, ou ["users", "orders"]
ignore_tables:                # Tabelas para ignorar
  - "schema_migrations"       # Rails/Laravel migrations
  - "ar_internal_metadata"    # Rails metadata
  - "goose_db_version"        # Goose migrations
  - "migrations"              # Generic migrations
  - "audit_logs"              # Audit/log tables
  - "sessions"                # Session data

# Opções de geração
template_dir: "./templates_postgresql"  # Templates otimizados
mock_provider: "testify"                    # "testify", "mock", ou "gomock"  
with_tests: true                           # Gerar testes (padrão: true)
```

#### pgx-goose-conf.json (Alternativa)
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

### Opções de Configuração Detalhadas

| Campo | Tipo | Padrão | Descrição |
|-------|------|--------|-----------|
| `dsn` | string | **obrigatório** | String de conexão PostgreSQL |
| `schema` | string | `"public"` | Schema do banco a introspeccionar |
| `output_dirs.base` | string | `"./pgx-goose"` | Diretório base para saída |
| `output_dirs.models` | string | `"{base}/models"` | Diretório para structs |
| `output_dirs.interfaces` | string | `"{base}/repository/interfaces"` | Diretório para interfaces |
| `output_dirs.repositories` | string | `"{base}/repository/postgres"` | Diretório para implementações |
| `output_dirs.mocks` | string | `"{base}/mocks"` | Diretório para mocks |
| `output_dirs.tests` | string | `"{base}/tests"` | Diretório para testes |
| `tables` | []string | `[]` (todas) | Lista de tabelas específicas |
| `ignore_tables` | []string | `[]` | Lista de tabelas a ignorar |
| `template_dir` | string | `""` (built-in) | Diretório de templates personalizados |
| `mock_provider` | string | `"testify"` | Provider de mocks: `testify`, `mock`, `gomock` |
| `with_tests` | bool | `true` | Gerar arquivos de teste |

### Validação e Regras

1. **DSN obrigatória**: O campo `dsn` é sempre obrigatório
2. **Conflito de tabelas**: Não é permitido especificar a mesma tabela em `tables` e `ignore_tables`
3. **Mock providers válidos**: Apenas `testify`, `mock` e `gomock` são aceitos
4. **Diretórios**: Se não especificados, usam padrões relativos ao `base`

### Precedência de Configuração

A configuração segue uma hierarquia de precedência (da maior para menor):

1. **CLI flags** (maior precedência)
2. **Arquivo de configuração** (`--config`)
3. **Valores padrão** (menor precedência)

```bash
# CLI sobrescreve qualquer valor do arquivo de configuração
pgx-goose --config pgx-goose-conf.yaml --schema "billing" --mock-provider "gomock"
```

### Filtragem de Tabelas

#### Modo Inclusivo (Tabelas Específicas)
```yaml
tables: ["users", "orders", "products"]  # Apenas essas tabelas
ignore_tables: []                        # Ignorar lista deve estar vazia
```

#### Modo Exclusivo (Todas Exceto...)
```yaml
tables: []  # Lista vazia = todas as tabelas
ignore_tables: 
  - "schema_migrations"      # Rails/Laravel
  - "ar_internal_metadata"   # ActiveRecord
  - "goose_db_version"       # Goose migrations
  - "audit_logs"             # Logs de auditoria
  - "sessions"               # Sessões de usuário
```

#### Validação de Conflitos
```yaml
# ❌ ERRO: Conflito detectado - tabela nas duas listas
tables: ["users", "orders"]
ignore_tables: ["users"]  # users aparece nas duas listas

# ✅ OK: Sem conflitos
tables: ["users", "orders"] 
ignore_tables: []
```

### Regras de Validação

O sistema aplica as seguintes validações antes da execução:

| Validação | Descrição | Erro |
|-----------|-----------|------|
| **DSN obrigatória** | Campo `dsn` deve estar presente | `DSN is required` |
| **Mock provider válido** | Deve ser `testify`, `mock` ou `gomock` | `invalid mock provider` |
| **Conflito de tabelas** | Tabela não pode estar em `tables` E `ignore_tables` | `conflicting table configuration` |
| **Arquivo de config** | Se especificado, deve existir e ser válido | `failed to read config file` |
| **Formato de config** | Deve ser `.yaml`, `.yml` ou `.json` | `unsupported config file format` |

#### Excluir Indesejadas (Recomendado)
```yaml
tables: []  # Todas as tabelas
ignore_tables: 
  - "schema_migrations"      # Rails/Laravel
  - "goose_db_version"       # Goose
  - "audit_logs"             # Logs
  - "sessions"               # Sessões temporárias
  - "cache"                  # Cache tables
```

### Configurações de Logging

```bash
# Logging padrão (apenas warnings/errors)
pgx-goose --config pgx-goose-conf.yaml

# Verbose (info + warnings + errors)
pgx-goose --config pgx-goose-conf.yaml --verbose

# Debug (tudo)
pgx-goose --config pgx-goose-conf.yaml --debug
```

## 📁 Estrutura Gerada

### Estrutura Padrão
```
generated/
├── models/                 # Structs das entidades
│   ├── user.go
│   └── product.go
├── repository/
│   ├── interfaces/         # Interfaces dos repositórios
│   │   ├── user_repository.go
│   │   └── product_repository.go
│   └── postgres/           # Implementações PostgreSQL
│       ├── user_repository.go
│       └── product_repository.go
├── mocks/                  # Mocks para testes
│   ├── mock_user_repository.go
│   └── mock_product_repository.go
└── tests/                  # Testes unitários/integração
    ├── user_repository_test.go
    └── product_repository_test.go
```

### Estrutura com Diretórios Personalizados
```
internal/
├── domain/entities/        # Models
├── ports/                  # Interfaces  
└── adapters/postgres/      # Implementações
tests/
├── mocks/                  # Mocks
└── integration/            # Testes
```

### Tipos de Arquivos Gerados

| Tipo | Descrição | Conteúdo |
|------|-----------|----------|
| **Models** | Structs das entidades | Tags JSON/DB, validação, métodos utilitários |
| **Interfaces** | Contratos dos repositórios | CRUD, transações, operações em lote |
| **Repositories** | Implementações PostgreSQL | Pool de conexões, prepared statements |
| **Mocks** | Testify/GoMock | Métodos de expectativa, assertions |
| **Tests** | Testes de integração | Setup/teardown, benchmarks, testcontainers |

## 🎨 Templates

### Templates Disponíveis

#### 1. Templates Padrão (`./templates/`)
- Templates genéricos para qualquer projeto Go
- Compatibilidade básica com pgx

#### 2. Templates PostgreSQL (`./templates_postgresql/`)
- **Recomendado** - Otimizados para `isis-golang-lib`
- Suporte a transações e operações em lote
- Métodos avançados nas structs

#### 3. Variações de Templates

Cada conjunto de templates possui duas variações:

| Template | Padrão | Simples (`*_simple.tmpl`) |
|----------|--------|-------------------------|
| **Model** | Struct completo com métodos utilitários | Struct básico apenas |
| **Repository** | Interface/implementação completa | Apenas operações CRUD básicas |
| **Mock** | Mock completo com todos os métodos | Mock simplificado |
| **Test** | Testes abrangentes com benchmarks | Testes unitários básicos |

### Usar Templates PostgreSQL
```bash
pgx-goose --template-dir "./templates_postgresql" --config pgx-goose-conf.yaml
```

### Templates Personalizados

Crie um diretório com:
```
my_templates/
├── model.tmpl                  # Structs
├── repository_interface.tmpl   # Interfaces
├── repository_postgres.tmpl    # Implementações
├── mock_testify.tmpl          # Mocks testify
├── mock_gomock.tmpl           # Mocks gomock
└── test.tmpl                  # Testes
```

**Exemplo model.tmpl:**
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

## � Referência da CLI

### Flags Principais

| Flag | Descrição | Valores | Padrão |
|------|-----------|---------|--------|
| `--dsn` | String de conexão PostgreSQL | `postgres://user:pass@host:port/db` | **obrigatório** |
| `--schema` | Schema do banco | `public`, `inventory`, `billing` | `public` |
| `--config` | Arquivo de configuração | `pgx-goose-conf.yaml`, `pgx-goose-conf.json` | - |
| `--out` | Diretório de saída | `./generated` | `./pgx-goose` |
| `--tables` | Tabelas específicas (CSV) | `users,orders,products` | todas |
| `--template-dir` | Diretório de templates | `./templates_postgresql` | built-in |
| `--mock-provider` | Provider de mocks | `testify`, `mock`, `gomock` | `testify` |
| `--with-tests` | Gerar testes | `true`, `false` | `true` |

### Flags de Diretórios Específicos

| Flag | Diretório | Exemplo |
|------|-----------|---------|
| `--models-dir` | Modelos/structs | `./internal/domain/entities` |
| `--interfaces-dir` | Interfaces | `./internal/ports` |
| `--repos-dir` | Implementações | `./internal/adapters/postgres` |
| `--mocks-dir` | Mocks | `./tests/mocks` |
| `--tests-dir` | Testes | `./tests/integration` |

### Flags de Configuração e Logging

| Flag | Descrição | Uso |
|------|-----------|-----|
| `--json` | Usar formato JSON para configuração | Para preferir .json ao .yaml |
| `--yaml` | Usar formato YAML para configuração | Padrão, explícito |
| `--verbose` | Log verboso (nível INFO) | Debug de execução |
| `--debug` | Log de debug (nível DEBUG) | Debug completo |

### Exemplos de Comandos

```bash
# Básico
pgx-goose --dsn "postgres://user:pass@localhost:5432/db"

# Schema customizado + tabelas específicas
pgx-goose --dsn "..." --schema "billing" --tables "invoices,payments"

# Configuração completa com logging
pgx-goose --config pgx-goose-conf.yaml --template-dir "./templates_postgresql" --verbose

# Ignorar tabelas específicas
pgx-goose --dsn "..." --ignore-tables "migrations,logs,sessions"

# Organização modular com diretórios customizados
pgx-goose --dsn "..." --tables "users" \
  --models-dir "./modules/user/entity" \
  --interfaces-dir "./modules/user/repository"

# Multi-schema empresarial
pgx-goose --schema "inventory" --out "./modules/inventory/generated"
pgx-goose --schema "billing" --out "./modules/billing/generated"

# Mock provider personalizado
pgx-goose --config pgx-goose-conf.yaml --mock-provider "gomock" --debug
```

## 💡 Exemplos de Uso

### 1. Projeto E-commerce Simples

**Schema SQL:**
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

**Configuração:**
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

**Gerar código:**
```bash
pgx-goose --config pgx-goose-conf.yaml
```

### 2. Arquitetura Multi-Schema Empresarial

```bash
# Schema de usuários
pgx-goose --schema "users" --out "./modules/users/generated"

# Schema de inventário
pgx-goose --schema "inventory" --out "./modules/inventory/generated"

# Schema de faturamento  
pgx-goose --schema "billing" --out "./modules/billing/generated"
```

### 3. Código Gerado - Exemplo User

**Model gerado (`models/user.go`):**
```go
type User struct {
    ID        int64     `json:"id" db:"id"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    Name      string    `json:"name" db:"name" validate:"required"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (u *User) TableName() string { return "users" }
func (u *User) Validate() error { /* validação */ }
func (u *User) Clone() *User { /* clonagem segura */ }
```

**Interface gerada (`interfaces/user_repository.go`):**
```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id int64) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id int64) error
    
    // Transações
    CreateTx(ctx context.Context, tx common.ITransaction, user *models.User) error
    
    // Operações em lote
    CreateBatch(ctx context.Context, users []*models.User) error
    
    // Buscas específicas
    FindByEmail(ctx context.Context, email string) (*models.User, error)
}
```

### 4. Usando o Código Gerado

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
    
    // Usar repositório gerado
    userRepo := postgres.NewUserRepository(pool)
    
    // Criar usuário
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
    fmt.Printf("User found: %+v\n", found)
}
```

### 5. Testando com Mocks

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

## 🤝 Contribuição

### Como Contribuir

1. **Fork** o projeto
2. **Crie uma branch** (`git checkout -b feature/NovaFeature`)
3. **Commit** suas mudanças (`git commit -m 'Add: nova feature'`)
4. **Push** para a branch (`git push origin feature/NovaFeature`)
5. **Abra um Pull Request**

### Desenvolvimento Local

```bash
# Clone e setup
git clone https://github.com/fsvxavier/isis-golang-lib.git
cd pgx-goose
go mod download

# Testes
go test ./...

# Build
go build -o pgx-goose .
./pgx-goose --help
```

### Estrutura do Projeto

```
pgx-goose/
├── cmd/                    # CLI commands (Cobra)
├── internal/
│   ├── config/            # Configuração
│   ├── generator/         # Geração de código
│   └── introspector/      # Introspecção PostgreSQL
├── templates/             # Templates padrão
├── templates_postgresql/  # Templates otimizados
├── examples/              # Exemplos de configuração
└── docs/                  # Documentação adicional
```

### Guidelines

- **Testes**: Toda nova funcionalidade deve ter testes
- **Documentação**: Atualize README.md para novas features
- **Templates**: Mantenha compatibilidade com templates existentes
- **Logs**: Use slog para logging estruturado

---

## 📄 Licença

Licenciado sob a [Licença MIT](LICENSE).

## 🙏 Agradecimentos

- [pgx](https://github.com/jackc/pgx) - Driver PostgreSQL de alta performance
- [Cobra](https://github.com/spf13/cobra) - Framework CLI
- [testify](https://github.com/stretchr/testify) - Framework de testes
- [testcontainers](https://github.com/testcontainers/testcontainers-go) - Testes de integração

## 📞 Suporte

- **Issues**: [GitHub Issues](https://github.com/fsvxavier/isis-golang-lib/issues)
- **Discussões**: [GitHub Discussions](https://github.com/fsvxavier/isis-golang-lib/discussions)

---

**pgx-goose** - Transformando seu PostgreSQL em código Go idiomático! 🚀

## 📚 Documentação

Documentação completa disponível em múltiplos idiomas:

- 🇧🇷 **[Português (Brasil)](docs/usage-pt-br.md)** - Documentação completa em português brasileiro
- 🇺🇸 **[English](docs/usage-en.md)** - Complete documentation in English  
- 🇪🇸 **[Español](docs/usage-es.md)** - Documentación completa en español
- 📋 **[Quick Reference](docs/quick-reference.md)** - Referência rápida de comandos e configurações

### O que está coberto na documentação:
- Instalação detalhada e pré-requisitos
- Configuração completa (YAML/JSON)
- Uso básico e avançado
- Exemplos práticos para diferentes cenários
- Estrutura de arquivos gerados
- Personalização com templates
- Troubleshooting e solução de problemas
- Integração com projetos (Makefile, CI/CD)

### Exemplos de Configuração
Veja a pasta [examples/](examples/) para:
- Configurações básicas e avançadas
- Setups específicos por ambiente (dev, prod, testing)
- Configurações para microserviços
- Exemplos de filtragem de tabelas
