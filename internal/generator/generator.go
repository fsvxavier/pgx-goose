package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/interfaces"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// Generator handles code generation with dependency injection
type Generator struct {
	config            *config.Config
	logger            interfaces.Logger
	metrics           interfaces.MetricsCollector
	templateOptimizer interfaces.TemplateOptimizer
	mu                sync.RWMutex
	generationStats   interfaces.GenerationMetrics
}

// New creates a new Generator with basic configuration
func New(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
		generationStats: interfaces.GenerationMetrics{
			ParallelWorkers: cfg.Parallel.Workers,
		},
	}
}

// NewWithDependencies creates a Generator with full dependency injection
func NewWithDependencies(
	cfg *config.Config,
	logger interfaces.Logger,
	metrics interfaces.MetricsCollector,
	templateOptimizer interfaces.TemplateOptimizer,
) interfaces.CodeGenerator {
	return &Generator{
		config:            cfg,
		logger:            logger,
		metrics:           metrics,
		templateOptimizer: templateOptimizer,
		generationStats: interfaces.GenerationMetrics{
			ParallelWorkers: cfg.Parallel.Workers,
		},
	}
}

// Generate generates code for the given schema
func (g *Generator) Generate(ctx context.Context, schema *introspector.Schema, outputPath string) error {
	start := time.Now()

	// Override output path if provided
	if outputPath != "" {
		g.config.OutputDir = outputPath
	}

	if g.logger != nil {
		g.logger.Info("Starting code generation",
			"tables", len(schema.Tables),
			"output", g.config.OutputDir)
	}

	// Reset stats
	g.mu.Lock()
	g.generationStats = interfaces.GenerationMetrics{
		ParallelWorkers: g.config.Parallel.Workers,
	}
	g.mu.Unlock()

	// Create output directories
	if err := g.createOutputDirectories(); err != nil {
		return fmt.Errorf("failed to create output directories: %w", err)
	}

	// Generate code based on configuration
	var err error
	if g.config.Parallel.Enabled && len(schema.Tables) > 1 {
		err = g.generateParallel(ctx, schema)
	} else {
		err = g.generateSequential(ctx, schema)
	}

	if err != nil {
		return err
	}

	// Update final metrics
	duration := time.Since(start).Seconds()
	g.mu.Lock()
	g.generationStats.Duration = duration
	g.generationStats.TablesProcessed = len(schema.Tables)
	g.mu.Unlock()

	if g.logger != nil {
		g.logger.Info("Code generation completed",
			"duration", duration,
			"tables", len(schema.Tables),
			"files", g.generationStats.FilesGenerated,
		)
	}

	// Record metrics
	if g.metrics != nil {
		g.metrics.RecordDuration("generation_duration", duration, map[string]string{
			"mode":   g.getGenerationMode(),
			"tables": fmt.Sprintf("%d", len(schema.Tables)),
		})
	}

	return nil
}

// generateParallel generates code using parallel workers
func (g *Generator) generateParallel(ctx context.Context, schema *introspector.Schema) error {
	workers := g.config.Parallel.Workers
	if workers <= 0 {
		workers = 4 // default
	}

	if g.logger != nil {
		g.logger.Info("Using parallel generation", "workers", workers)
	}

	// Create work channel
	tableChan := make(chan introspector.Table, len(schema.Tables))
	errorChan := make(chan error, len(schema.Tables))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go g.worker(ctx, i, tableChan, errorChan, &wg)
	}

	// Send work
	for _, table := range schema.Tables {
		select {
		case tableChan <- table:
		case <-ctx.Done():
			close(tableChan)
			return ctx.Err()
		}
	}
	close(tableChan)

	// Wait for completion
	wg.Wait()
	close(errorChan)

	// Check for errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// worker is a parallel worker for table processing
func (g *Generator) worker(ctx context.Context, id int, tables <-chan introspector.Table, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	if g.logger != nil {
		g.logger.Debug("Worker started", "id", id)
	}

	for {
		select {
		case table, ok := <-tables:
			if !ok {
				// Channel closed, worker done
				if g.logger != nil {
					g.logger.Debug("Worker completed", "id", id)
				}
				errors <- nil
				return
			}

			if err := g.generateTableFiles(table); err != nil {
				if g.logger != nil {
					g.logger.Error("Worker failed to generate table",
						"worker", id,
						"table", table.Name,
						"error", err)
				}
				errors <- fmt.Errorf("worker %d failed on table %s: %w", id, table.Name, err)
				return
			}

			g.mu.Lock()
			g.generationStats.FilesGenerated += 4 // model, interface, repo, test
			g.mu.Unlock()

		case <-ctx.Done():
			// Context cancelled
			if g.logger != nil {
				g.logger.Debug("Worker cancelled", "id", id)
			}
			errors <- ctx.Err()
			return
		}
	}
}

// generateSequential generates code sequentially
func (g *Generator) generateSequential(ctx context.Context, schema *introspector.Schema) error {
	if g.logger != nil {
		g.logger.Info("Using sequential generation")
	}

	for _, table := range schema.Tables {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := g.generateTableFiles(table); err != nil {
			g.mu.Lock()
			g.generationStats.ErrorsCount++
			g.mu.Unlock()
			return fmt.Errorf("failed to generate files for table %s: %w", table.Name, err)
		}

		g.mu.Lock()
		g.generationStats.FilesGenerated += 4 // model, interface, repo, test
		g.mu.Unlock()
	}

	return nil
}

// generateTableFiles generates all files for a table
func (g *Generator) generateTableFiles(table introspector.Table) error {
	// Generate model
	if err := g.generateModel(table); err != nil {
		return fmt.Errorf("failed to generate model: %w", err)
	}

	// Generate repository interface
	if err := g.generateRepositoryInterface(table); err != nil {
		return fmt.Errorf("failed to generate repository interface: %w", err)
	}

	// Generate repository implementation
	if err := g.generateRepository(table); err != nil {
		return fmt.Errorf("failed to generate repository: %w", err)
	}

	// Generate tests if enabled
	if g.config.WithTests {
		if err := g.generateTests(table); err != nil {
			return fmt.Errorf("failed to generate tests: %w", err)
		}
	}

	return nil
}

// createOutputDirectories creates necessary output directories
func (g *Generator) createOutputDirectories() error {
	baseDir := g.config.OutputDir
	if baseDir == "" {
		baseDir = "./generated"
	}

	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "models"),
		filepath.Join(baseDir, "interfaces"),
		filepath.Join(baseDir, "repositories"),
		filepath.Join(baseDir, "tests"),
		filepath.Join(baseDir, "mocks"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateModel generates the model file for a table
func (g *Generator) generateModel(table introspector.Table) error {
	template := g.getModelTemplate()
	data := map[string]interface{}{
		"Table":     table,
		"TableName": toPascalCase(table.Name),
		"Package":   "models",
	}

	content, err := g.executeTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to execute model template: %w", err)
	}

	filename := filepath.Join(g.config.OutputDir, "models", strings.ToLower(table.Name)+".go")
	return os.WriteFile(filename, []byte(content), 0644)
}

// generateRepositoryInterface generates the repository interface
func (g *Generator) generateRepositoryInterface(table introspector.Table) error {
	template := g.getRepositoryInterfaceTemplate()
	data := map[string]interface{}{
		"Table":     table,
		"TableName": toPascalCase(table.Name),
		"Package":   "interfaces",
	}

	content, err := g.executeTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to execute repository interface template: %w", err)
	}

	filename := filepath.Join(g.config.OutputDir, "interfaces", strings.ToLower(table.Name)+"_repository.go")
	return os.WriteFile(filename, []byte(content), 0644)
}

// generateRepository generates the repository implementation
func (g *Generator) generateRepository(table introspector.Table) error {
	template := g.getRepositoryTemplate()
	data := map[string]interface{}{
		"Table":     table,
		"TableName": toPascalCase(table.Name),
		"Package":   "repositories",
	}

	content, err := g.executeTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to execute repository template: %w", err)
	}

	filename := filepath.Join(g.config.OutputDir, "repositories", strings.ToLower(table.Name)+"_repository.go")
	return os.WriteFile(filename, []byte(content), 0644)
}

// generateTests generates test files
func (g *Generator) generateTests(table introspector.Table) error {
	template := g.getTestTemplate()
	data := map[string]interface{}{
		"Table":     table,
		"TableName": toPascalCase(table.Name),
		"Package":   "tests",
	}

	content, err := g.executeTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to execute test template: %w", err)
	}

	filename := filepath.Join(g.config.OutputDir, "tests", strings.ToLower(table.Name)+"_test.go")
	return os.WriteFile(filename, []byte(content), 0644)
}

// executeTemplate executes a template with the given data
func (g *Generator) executeTemplate(templateStr string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"toPascalCase": toPascalCase,
		"toLower":      strings.ToLower,
		"add": func(a, b int) int {
			return a + b
		},
	}

	tmpl, err := template.New("generator").Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// SetTemplateOptimizer sets the template optimizer
func (g *Generator) SetTemplateOptimizer(optimizer interfaces.TemplateOptimizer) {
	g.templateOptimizer = optimizer
}

// GetMetrics returns current generation metrics
func (g *Generator) GetMetrics() interfaces.GenerationMetrics {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Add cache stats if available
	if g.templateOptimizer != nil {
		cacheStats := g.templateOptimizer.GetCacheStats()
		g.generationStats.CacheHitRatio = cacheStats.HitRatio
	}

	return g.generationStats
}

// getGenerationMode returns the current generation mode
func (g *Generator) getGenerationMode() string {
	if g.config.Parallel.Enabled {
		return "parallel"
	}
	return "sequential"
}

// getModelTemplate returns the model template
func (g *Generator) getModelTemplate() string {
	return `package {{.Package}}

import (
	"time"
)

// {{.TableName}} represents the {{.Table.Name}} table
type {{.TableName}} struct {
{{- range .Table.Columns}}
	{{toPascalCase .Name}} {{.GoType}} ` + "`json:\"{{.Name}}\" db:\"{{.Name}}\"`" + `{{if .Comment}} // {{.Comment}}{{end}}
{{- end}}
}

// TableName returns the table name
func ({{.TableName}}) TableName() string {
	return "{{.Table.Name}}"
}
`
}

// getRepositoryInterfaceTemplate returns the repository interface template
func (g *Generator) getRepositoryInterfaceTemplate() string {
	return `package {{.Package}}

import (
	"context"
)

// {{.TableName}}Repository defines the interface for {{.Table.Name}} operations
type {{.TableName}}Repository interface {
	Create(ctx context.Context, entity interface{}) error
	GetByID(ctx context.Context, id interface{}) (interface{}, error)
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id interface{}) error
	List(ctx context.Context, limit, offset int) ([]interface{}, error)
}
`
}

// getRepositoryTemplate returns the repository implementation template
func (g *Generator) getRepositoryTemplate() string {
	return `package {{.Package}}

import (
	"context"
	"fmt"
)

// {{.TableName}}Repository implements the {{.TableName}}Repository interface
type {{.TableName}}Repository struct {
	// Add database connection field here
}

// New{{.TableName}}Repository creates a new {{.TableName}}Repository
func New{{.TableName}}Repository() *{{.TableName}}Repository {
	return &{{.TableName}}Repository{}
}

// Create creates a new {{.Table.Name}} record
func (r *{{.TableName}}Repository) Create(ctx context.Context, entity interface{}) error {
	return fmt.Errorf("not implemented")
}

// GetByID retrieves a {{.Table.Name}} by ID
func (r *{{.TableName}}Repository) GetByID(ctx context.Context, id interface{}) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// Update updates a {{.Table.Name}} record
func (r *{{.TableName}}Repository) Update(ctx context.Context, entity interface{}) error {
	return fmt.Errorf("not implemented")
}

// Delete deletes a {{.Table.Name}} record
func (r *{{.TableName}}Repository) Delete(ctx context.Context, id interface{}) error {
	return fmt.Errorf("not implemented")
}

// List retrieves a list of {{.Table.Name}} records
func (r *{{.TableName}}Repository) List(ctx context.Context, limit, offset int) ([]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
`
}

// getTestTemplate returns the test template
func (g *Generator) getTestTemplate() string {
	return `package {{.Package}}

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func Test{{.TableName}}Repository_Create(t *testing.T) {
	t.Skip("Implementation pending")
}

func Test{{.TableName}}Repository_GetByID(t *testing.T) {
	t.Skip("Implementation pending")
}

func Test{{.TableName}}Repository_Update(t *testing.T) {
	t.Skip("Implementation pending")
}

func Test{{.TableName}}Repository_Delete(t *testing.T) {
	t.Skip("Implementation pending")
}

func Test{{.TableName}}Repository_List(t *testing.T) {
	t.Skip("Implementation pending")
}
`
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return result
}
