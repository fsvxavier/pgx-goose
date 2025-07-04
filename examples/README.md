# Exemplos de Configuração pgx-goose

Esta pasta contém diferentes exemplos de arquivos de configuração para o pgx-goose, demonstrando várias abordagens e cenários de uso.

## Arquivos Disponíveis

### Configurações Básicas
- **`pgx-goose-conf_basic.yaml`** - Configuração simples e direta para começar rapidamente
- **`pgx-goose-conf_basic.json`** - Mesma configuração básica em formato JSON

### Configurações Avançadas
- **`pgx-goose-conf_advanced.yaml`** - Configuração completa com diretórios separados e todas as opções
- **`pgx-goose-conf_separate_dirs.yaml`** - Foco na organização com diretórios separados por tipo

### Configurações por Ambiente
- **`pgx-goose-conf_development.yaml`** - Otimizada para desenvolvimento local
- **`pgx-goose-conf_production.yaml`** - Configuração robusta para produção
- **`pgx-goose-conf_testing.yaml`** - Para testes automatizados e CI/CD

### Configurações por Arquitetura
- **`pgx-goose-conf_microservice.yaml`** - Para projetos de microserviços
- **`pgx-goose-conf_custom_schema.yaml`** - Para trabalhar com schemas específicos

### Configurações por Filtragem
- **`pgx-goose-conf_ignore_tables.yaml`** - Exemplo de como ignorar tabelas específicas

## Como Usar

1. **Copie** o arquivo de exemplo que melhor se adequa ao seu projeto
2. **Renomeie** para `pgx-goose-conf.yaml` ou `pgx-goose-conf.json`
3. **Edite** as configurações específicas do seu projeto:
   - DSN do banco de dados
   - Schema
   - Diretórios de saída
   - Tabelas específicas ou a ignorar

## Exemplos de Uso

### Uso com arquivo de configuração específico:
```bash
pgx-goose --config examples/pgx-goose-conf_basic.yaml
```

### Uso com busca automática (renomeie o arquivo):
```bash
cp examples/pgx-goose-conf_basic.yaml pgx-goose-conf.yaml
pgx-goose
```

## Estrutura dos Arquivos de Configuração

### Campos Principais:
- **`dsn`** - String de conexão PostgreSQL
- **`schema`** - Schema do banco a ser processado (padrão: "public")
- **`out`** - Diretório de saída simples (legado)
- **`output_dirs`** - Configuração detalhada de diretórios
- **`mock_provider`** - Provider de mocks ("testify" ou "mock")
- **`with_tests`** - Se deve gerar testes (true/false)
- **`template_dir`** - Diretório de templates customizados (opcional)
- **`tables`** - Lista de tabelas específicas (vazio = todas)
- **`ignore_tables`** - Lista de tabelas a ignorar

### Configuração de Diretórios (output_dirs):
- **`base`** - Diretório base
- **`models`** - Entidades/modelos
- **`interfaces`** - Interfaces dos repositórios
- **`repositories`** - Implementações PostgreSQL
- **`mocks`** - Mocks para testes
- **`tests`** - Testes de integração

## Dicas

1. **Ambiente de Desenvolvimento**: Use configurações mais simples e rápidas
2. **Produção**: Use todas as validações e testes
3. **Microserviços**: Foque em schemas específicos
4. **CI/CD**: Use configurações otimizadas para testes automatizados
5. **Clean Architecture**: Organize os diretórios conforme a estrutura do seu projeto

## Variáveis de Ambiente

Você pode usar variáveis de ambiente no DSN:
```yaml
dsn: "postgres://user:${DB_PASSWORD}@${DB_HOST}:5432/mydb"
```
