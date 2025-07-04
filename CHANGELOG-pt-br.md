# Registro de Mudanças

Todas as mudanças notáveis deste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
e este projeto adere ao [Versionamento Semântico](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-03

### 🎯 Funcionalidades Principais Implementadas

#### ✅ Diretórios de Saída Configuráveis (NOVA FUNCIONALIDADE)
- Adicionada configuração `output_dirs` para diretórios separados
- Suporte a flags CLI específicas (`--models-dir`, `--interfaces-dir`, etc.)
- Retrocompatibilidade com configuração `OutputDir` legacy
- Precedência: CLI flags > config file > defaults
- Suporte a 5 arquiteturas diferentes (Hexagonal, Clean, DDD, Modular, Monorepo)

#### ✅ Templates PostgreSQL Otimizados (NOVA FUNCIONALIDADE)
- Templates especializados na pasta `templates_postgresql/`
- Integração com provider `db/postgresql` da nexs-lib
- Suporte a operações transacionais e em lote
- Métodos avançados nas entidades (TableName, Clone, Validate, etc.)
- Soft delete quando aplicável

#### ✅ Documentação Unificada (NOVA FUNCIONALIDADE)
- README.md completo com ~1.315 linhas
- 17 seções principais com 15+ exemplos completos
- Integração de 6 arquivos de documentação
- Índice detalhado para navegação
- Casos de uso por arquitetura

### Adicionado

#### Core Features
- ✅ Ferramenta CLI completa baseada em Cobra
- ✅ Introspecção completa de schemas PostgreSQL
- ✅ Geração automática de structs Go a partir de tabelas
- ✅ Geração de interfaces de repositórios com CRUD completo
- ✅ Implementações PostgreSQL usando pgx/v5
- ✅ Suporte a dois providers de mock: testify e gomock
- ✅ Geração automática de testes unitários
- ✅ Sistema de templates customizáveis usando Go Templates

#### Configuration & CLI
- ✅ Suporte a arquivos de configuração YAML e JSON
- ✅ Flags de linha de comando abrangentes
- ✅ Sistema de logging configurável (debug, verbose, info, warn, error)
- ✅ Validação de configuração robusta

#### Database Support
- ✅ Introspecção completa de tabelas, colunas, tipos
- ✅ Suporte a chaves primárias, índices e chaves estrangeiras
- ✅ Mapeamento automático de tipos PostgreSQL → Go
- ✅ Suporte a tipos nullable com ponteiros
- ✅ Comentários de tabelas e colunas preservados

#### Code Generation
- ✅ Templates embutidos para todos os tipos de arquivo
- ✅ Estrutura de projeto organizada e idiomática
- ✅ Suporte a templates personalizados via diretório customizado
- ✅ Geração de código limpo seguindo convenções Go
- ✅ Suporte a relacionamentos entre tabelas

#### Testing & Quality
- ✅ Testes unitários abrangentes
- ✅ Mocks automáticos para todas as interfaces
- ✅ Testes gerados com cenários de sucesso e erro
- ✅ Integração com testify/assert e testify/mock
- ✅ Suporte a gomock para projetos que preferem

### Características Técnicas

#### Arquitetura
- Clean Architecture com separação de responsabilidades
- Design orientado a interfaces
- Injeção de dependências
- Estrutura modular e extensível

#### Tipos Suportados
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

#### Estrutura de Projeto Gerada
```
output_dir/
├── models/                     # Structs das tabelas
├── repository/
│   ├── interfaces/             # Interfaces dos repositórios
│   └── postgres/               # Implementações PostgreSQL
├── mocks/                      # Mocks para testes
└── tests/                      # Testes unitários
```

#### Flags CLI Disponíveis
- `--dsn`: String de conexão PostgreSQL
- `--out`: Diretório de saída
- `--tables`: Lista de tabelas específicas
- `--config`: Arquivo de configuração
- `--template-dir`: Templates customizados
- `--mock-provider`: Provider de mocks (testify/mock)
- `--with-tests`: Geração de testes
- `--verbose`: Logging verboso
- `--debug`: Logging de debug

### Arquivos de Projeto

#### Documentação
- ✅ README.md completo com exemplos
- ✅ EXAMPLES.md com casos de uso detalhados
- ✅ Templates de configuração (YAML/JSON)
- ✅ Scripts de demonstração (Bash/PowerShell)

#### Build & Development
- ✅ Makefile com targets úteis
- ✅ go.mod com todas as dependências
- ✅ .gitignore apropriado
- ✅ Licença MIT

#### Scripts & Tools
- ✅ demo.sh para Linux/macOS
- ✅ demo.ps1 para Windows
- ✅ Schema SQL de exemplo
- ✅ Configurações de exemplo

### Dependências

#### Runtime
- `github.com/jackc/pgx/v5` - Driver PostgreSQL
- `github.com/spf13/cobra` - Framework CLI
- `log/slog` - Structured logging (native Go 1.21+)
- `gopkg.in/yaml.v3` - Parser YAML

#### Testing & Mocking
- `github.com/stretchr/testify` - Framework de testes
- `go.uber.org/mock` - Gomock para mocks
- `github.com/google/uuid` - Suporte a UUID
- `github.com/shopspring/decimal` - Tipos decimais

### Notas de Uso

#### Instalação
```bash
go install github.com/fsvxavier/pgx-goose@latest
```

#### Uso Básico
```bash
pgx-goose --dsn "postgres://user:pass@localhost:5432/db" --out ./generated
```

#### Com Configuração
```bash
pgx-goose --config pgx-goose-conf.yaml --verbose
```

### Agradecimentos

Este projeto foi inspirado por:
- [xo/dbtpl](https://github.com/xo/dbtpl)
- [go-gorm/gen](https://github.com/go-gorm/gen)
- Princípios de Clean Architecture
- Padrões SOLID e DDD

### Contribuindo

Para contribuir com o projeto:
1. Fork o repositório
2. Crie uma branch para sua feature
3. Implemente com testes
4. Execute os testes e linting
5. Submeta um Pull Request

### Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.
