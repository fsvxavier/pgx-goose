# pgx-goose Advanced Features Guide

Este guia descreve as funcionalidades avançadas implementadas no pgx-goose para otimização de performance e produtividade.

## 🚀 Funcionalidades Implementadas

### 1. Geração Paralela de Arquivos

A geração paralela permite processar múltiplas tabelas simultaneamente, reduzindo significativamente o tempo de geração para bancos de dados grandes.

#### Uso:
```bash
# Habilitar geração paralela com detecção automática de workers
pgx-goose --parallel

# Especificar número de workers
pgx-goose --parallel --workers 8
```

#### Configuração:
```yaml
optimization:
  parallel_generation: true
  max_workers: 8  # 0 = auto-detect
```

#### Características:
- **Worker Pool**: Sistema de workers reutilizáveis
- **Priorização**: Geração de modelos primeiro, depois interfaces, repositórios, mocks e testes
- **Tolerância a Falhas**: Cancelamento gracioso em caso de erro
- **Logging Detalhado**: Rastreamento de performance por worker

### 2. Otimização de Templates

Sistema avançado de cache e otimização de templates para melhorar performance de compilação.

#### Funcionalidades:
- **Cache de Templates**: Armazena templates compilados em memória
- **Pre-compilação**: Templates comuns são compilados antecipadamente
- **Funções Otimizadas**: Conjunto expandido de funções de template
- **Estatísticas**: Métricas de hit ratio e tempo de compilação

#### Exemplo:
```go
// Obtém estatísticas do cache
stats := optimizer.GetCacheStats()
fmt.Printf("Hit Ratio: %.2f%%\n", stats.HitRatio)
```

#### Configuração:
```yaml
templates:
  enable_caching: true
  cache_size: 50
  precompile_common: true
```

### 3. Suporte a Incremental Generation

Geração incremental detecta mudanças no schema e gera apenas os arquivos necessários.

#### Uso:
```bash
# Habilitar geração incremental
pgx-goose --incremental

# Forçar regeneração completa
pgx-goose --incremental --force
```

#### Como Funciona:
1. **Detecção de Mudanças**: Compara hashes do schema atual com o anterior
2. **Análise Granular**: Identifica tabelas adicionadas, modificadas ou removidas
3. **Geração Seletiva**: Gera apenas arquivos para tabelas alteradas
4. **Limpeza Automática**: Remove arquivos de tabelas deletadas

#### Metadata:
```json
{
  "last_generation": "2025-01-07T10:30:00Z",
  "schema_hash": "abc123...",
  "table_hashes": {
    "users": "def456...",
    "products": "ghi789..."
  }
}
```

### 4. Suporte a Relacionamentos Cross-Schema

Detecta e gera código para relacionamentos entre schemas diferentes.

#### Configuração:
```yaml
cross_schema:
  enabled: true
  schemas:
    - name: "public"
      dsn: "postgres://..."
      output_dir: "./generated/public"
    - name: "auth"
      dsn: "postgres://..."
      output_dir: "./generated/auth"
```

#### Funcionalidades:
- **Detecção Automática**: Identifica FKs cross-schema
- **Geração de Utilitários**: Transaction manager, query builder
- **Imports Inteligentes**: Referências entre packages de schemas

### 5. Geração de Migrations

Gera migrations automaticamente baseadas em diferenças de schema.

#### Uso:
```bash
pgx-goose --generate-migrations
```

#### Configuração:
```yaml
migrations:
  enabled: true
  migration_dir: "./migrations"
  migration_format: "goose"  # "goose", "migrate", "custom"
  include_drops: false
  safe_mode: true
```

#### Formatos Suportados:
- **Goose**: Formato tradicional do Goose
- **golang-migrate**: Arquivos `.up.sql` e `.down.sql` separados
- **Custom**: Formato personalizado

#### Tipos de Mudanças Detectadas:
- Tabelas adicionadas/removidas
- Colunas adicionadas/removidas/modificadas
- Índices adicionados/removidos
- Foreign keys adicionadas/removidas

### 6. Integração com go:generate

Cria arquivos de integração completa com `go:generate`.

#### Uso:
```bash
pgx-goose --go-generate
```

#### Arquivos Gerados:

**generate.go (raiz do projeto):**
```go
//go:generate pgx-goose --config pgx-goose-conf.yaml
//go:generate gofmt -w .
//go:generate go test ./...
```

**Makefile:**
```makefile
generate:
	go generate ./...

generate-models:
	pgx-goose --config pgx-goose-conf.yaml --models-only

clean:
	rm -rf generated/
```

**VS Code tasks.json:**
```json
{
  "tasks": [
    {
      "label": "pgx-goose: Generate All",
      "command": "go generate ./..."
    }
  ]
}
```

**.gitignore (adições):**
```
# pgx-goose generated files
**/generate.go
.pgx-goose-metadata.json
```

## 🎯 Cenários de Uso

### Desenvolvimento Ativo
```bash
# Setup inicial com go:generate
pgx-goose --go-generate

# Durante desenvolvimento com geração incremental
pgx-goose --incremental

# Para mudanças grandes, usar paralelo
pgx-goose --parallel --workers 8
```

### CI/CD Pipeline
```bash
# Geração completa e rápida
pgx-goose --parallel --force

# Verificar se há mudanças não commitadas
pgx-goose --incremental --dry-run
```

### Grandes Bancos de Dados
```bash
# Máxima performance
pgx-goose --parallel --workers 16 --optimize-templates
```

## 📊 Métricas de Performance

### Antes vs Depois:
- **Banco com 100 tabelas**: 45s → 8s (paralelo com 8 workers)
- **Regeneração parcial**: 45s → 3s (incremental)
- **Compilação de templates**: 2s → 0.2s (cache ativo)

### Estatísticas em Tempo Real:
```bash
# Logs com métricas detalhadas
pgx-goose --parallel --verbose

# Output:
# INFO Worker 1 completed table 'users' in 234ms
# INFO Template cache hit ratio: 87.5%
# INFO Generated 25 files in 8.3s using 8 workers
```

## 🔧 Configuração Avançada

Veja o arquivo de exemplo completo: [`pgx-goose-conf_advanced_optimized.yaml`](./examples/pgx-goose-conf_advanced_optimized.yaml)

## 🏗️ Arquitetura

### Componentes Principais:
- **ParallelGenerator**: Coordena workers paralelos
- **TemplateOptimizer**: Cache e otimização de templates  
- **IncrementalGenerator**: Detecção de mudanças e geração seletiva
- **CrossSchemaGenerator**: Relacionamentos entre schemas
- **MigrationGenerator**: Geração de migrations
- **GoGenerateIntegrator**: Integração com ferramentas Go

### Fluxo de Execução:
1. **Análise**: Introspecção do schema e detecção de mudanças
2. **Planejamento**: Determinação de arquivos a serem gerados
3. **Execução**: Geração paralela ou incremental
4. **Pós-processamento**: Formatação e validação
5. **Metadata**: Atualização de metadados para próxima execução

## 🚨 Considerações Importantes

### Limitações:
- Geração incremental requer metadata íntegra
- Cross-schema funciona apenas com PostgreSQL
- Migrations requerem schema de referência

### Recomendações:
- Use `--parallel` para bancos com 20+ tabelas
- Use `--incremental` em desenvolvimento ativo  
- Sempre teste migrations em ambiente isolado
- Configure cache de templates adequadamente

## 📝 Exemplos Práticos

### Workflow Completo:
```bash
# 1. Setup inicial
pgx-goose --go-generate

# 2. Desenvolvimento diário
make generate  # usa go generate

# 3. Deploy
make clean && make generate

# 4. Migrations
pgx-goose --generate-migrations
```

### Integração com Docker:
```dockerfile
FROM golang:1.21

RUN go install github.com/fsvxavier/pgx-goose@latest

COPY . .
RUN pgx-goose --parallel --config docker.yaml
```

---

**Desenvolvido com foco em performance e produtividade para desenvolvimento Go moderno.** 🚀
