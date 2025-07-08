# Dependency Injection Example

Este exemplo demonstra o uso do container de injeção de dependências implementado no pgx-goose.

## 🎯 Objetivos

- Demonstrar o padrão de Dependency Injection
- Mostrar o gerenciamento centralizado de dependências
- Exemplificar testes com componentes desacoplados
- Apresentar as métricas e observabilidade

## 🏗️ Arquitetura

```
Container
├── Configuration (Config)
├── Logger (Structured)
├── Metrics (Collector)
├── Database Pool (PGX)
├── Schema Introspector
├── Template Optimizer
└── Code Generator
```

## 🔄 Fluxo de Dependências

1. **Config** → Carrega configuração
2. **Logger** → Sistema de logs estruturados
3. **Metrics** → Coleta de métricas
4. **Database** → Pool de conexões otimizado
5. **Introspector** → Análise do schema
6. **Optimizer** → Cache de templates
7. **Generator** → Geração de código

## 📊 Benefícios

### ✅ Testabilidade
- Injeção de mocks simples
- Isolamento de componentes
- Testes unitários focados

### ✅ Manutenibilidade
- Baixo acoplamento
- Alta coesão
- Responsabilidades bem definidas

### ✅ Performance
- Pool de conexões otimizado
- Cache de templates
- Geração paralela

### ✅ Observabilidade
- Logs estruturados
- Métricas detalhadas
- Health checks

## 🚀 Como Executar

```bash
# Executar o exemplo
go run main.go

# Executar com configuração personalizada
DATABASE_URL="postgres://user:pass@localhost:5432/db" go run main.go
```

## 📋 Exemplo de Uso

```go
// 1. Criar configuração
config := &config.Config{
    DSN: "postgres://user:pass@localhost:5432/db",
    Schema: "public",
    OutputDir: "./generated",
    WithTests: true,
    Parallel: config.ParallelConfig{
        Enabled: true,
        Workers: 4,
    },
}

// 2. Criar container
container, err := container.NewContainer(config)
if err != nil {
    return err
}
defer container.Close()

// 3. Usar componentes
logger := container.GetLogger()
generator := container.GetGenerator()

// 4. Gerar código
ctx := context.Background()
err = generator.Generate(ctx, schema, outputPath)
```

## 🧪 Testes

```go
func TestWithContainer(t *testing.T) {
    // Criar container com mocks
    testConfig := &config.Config{
        DSN: "mock://test",
        // ... configuração de teste
    }
    
    container, err := container.NewContainer(testConfig)
    require.NoError(t, err)
    defer container.Close()
    
    // Testar componentes individualmente
    logger := container.GetLogger()
    assert.NotNil(t, logger)
    
    metrics := container.GetMetrics()
    assert.NotNil(t, metrics)
    
    // Testar health check
    err = container.Health(context.Background())
    assert.NoError(t, err)
}
```

## 📈 Métricas Disponíveis

- `generation_duration` - Duração da geração
- `templates_compiled` - Templates compilados
- `cache_hit_ratio` - Taxa de hit do cache
- `tables_processed` - Tabelas processadas
- `parallel_workers` - Workers paralelos

## 🔍 Health Checks

O container implementa health checks para:
- Conectividade com banco de dados
- Status dos componentes internos
- Validação da configuração

## 💡 Boas Práticas

1. **Sempre fechar o container**: Use `defer container.Close()`
2. **Validar erros**: Sempre verificar erros na criação
3. **Usar contexto**: Passar contexto para operações longas
4. **Monitorar métricas**: Acompanhar performance
5. **Logs estruturados**: Usar o logger do container
