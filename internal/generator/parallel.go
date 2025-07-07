package generator

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/config"
	"github.com/fsvxavier/pgx-goose/internal/introspector"
)

// GenerationTask represents a task for generating code
type GenerationTask struct {
	Type     GenerationType
	Table    introspector.Table
	Template string
	Output   string
	Priority int // Lower values = higher priority
}

// GenerationType represents the type of code generation
type GenerationType int

const (
	ModelGeneration GenerationType = iota
	InterfaceGeneration
	RepositoryGeneration
	MockGeneration
	TestGeneration
)

// ParallelGenerator handles parallel code generation
type ParallelGenerator struct {
	*Generator
	maxWorkers int
	workerPool chan struct{}
	taskQueue  chan GenerationTask
	results    chan GenerationResult
	errorChan  chan error
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// GenerationResult represents the result of a generation task
type GenerationResult struct {
	Task     GenerationTask
	Success  bool
	Error    error
	Duration string
}

// NewParallelGenerator creates a new parallel generator
func NewParallelGenerator(cfg *config.Config, maxWorkers int) *ParallelGenerator {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pg := &ParallelGenerator{
		Generator:  New(cfg),
		maxWorkers: maxWorkers,
		workerPool: make(chan struct{}, maxWorkers),
		taskQueue:  make(chan GenerationTask, 100),
		results:    make(chan GenerationResult, 100),
		errorChan:  make(chan error, 10),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Initialize worker pool
	for i := 0; i < maxWorkers; i++ {
		pg.workerPool <- struct{}{}
	}

	return pg
}

// GenerateParallel generates code using parallel workers
func (pg *ParallelGenerator) GenerateParallel(schema *introspector.Schema) error {
	slog.Info("Starting parallel code generation", "workers", pg.maxWorkers, "tables", len(schema.Tables))

	// Create output directory structure first
	if err := pg.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Start result collector
	go pg.collectResults()

	// Start workers
	for i := 0; i < pg.maxWorkers; i++ {
		pg.wg.Add(1)
		go pg.worker(i)
	}

	// Queue tasks with priorities
	go func() {
		defer close(pg.taskQueue)
		pg.queueTasks(schema)
	}()

	// Wait for all workers to complete
	pg.wg.Wait()
	close(pg.results)
	close(pg.errorChan)

	// Check for errors
	select {
	case err := <-pg.errorChan:
		return err
	default:
		slog.Info("Parallel code generation completed successfully")
		return nil
	}
}

// worker processes generation tasks
func (pg *ParallelGenerator) worker(id int) {
	defer pg.wg.Done()

	slog.Debug("Worker started", "worker_id", id)

	for {
		select {
		case <-pg.ctx.Done():
			slog.Debug("Worker cancelled", "worker_id", id)
			return
		case task, ok := <-pg.taskQueue:
			if !ok {
				slog.Debug("Worker finished - no more tasks", "worker_id", id)
				return
			}

			// Acquire worker slot
			<-pg.workerPool

			result := pg.processTask(task, id)
			pg.results <- result

			// Release worker slot
			pg.workerPool <- struct{}{}

			if !result.Success {
				select {
				case pg.errorChan <- result.Error:
				default:
					// Error channel is full, cancel context
					pg.cancel()
				}
				return
			}
		}
	}
}

// processTask processes a single generation task
func (pg *ParallelGenerator) processTask(task GenerationTask, workerID int) GenerationResult {
	slog.Debug("Processing task",
		"worker_id", workerID,
		"table", task.Table.Name,
		"type", task.Type,
		"priority", task.Priority)

	start := time.Now()

	var err error
	switch task.Type {
	case ModelGeneration:
		err = pg.generateSingleModel(task.Table)
	case InterfaceGeneration:
		err = pg.generateSingleInterface(task.Table)
	case RepositoryGeneration:
		err = pg.generateSingleRepository(task.Table)
	case MockGeneration:
		err = pg.generateSingleMock(task.Table)
	case TestGeneration:
		err = pg.generateSingleTest(task.Table)
	default:
		err = fmt.Errorf("unknown generation type: %d", task.Type)
	}

	duration := time.Since(start)

	return GenerationResult{
		Task:     task,
		Success:  err == nil,
		Error:    err,
		Duration: duration.String(),
	}
}

// queueTasks queues all generation tasks with priorities
func (pg *ParallelGenerator) queueTasks(schema *introspector.Schema) {
	// Priority order: Models (1) -> Interfaces (2) -> Repositories (3) -> Mocks (4) -> Tests (5)

	// Queue model generation first (highest priority)
	for _, table := range schema.Tables {
		task := GenerationTask{
			Type:     ModelGeneration,
			Table:    table,
			Priority: 1,
		}
		pg.taskQueue <- task
	}

	// Queue interface generation
	for _, table := range schema.Tables {
		task := GenerationTask{
			Type:     InterfaceGeneration,
			Table:    table,
			Priority: 2,
		}
		pg.taskQueue <- task
	}

	// Queue repository generation
	for _, table := range schema.Tables {
		task := GenerationTask{
			Type:     RepositoryGeneration,
			Table:    table,
			Priority: 3,
		}
		pg.taskQueue <- task
	}

	// Queue mock generation
	for _, table := range schema.Tables {
		task := GenerationTask{
			Type:     MockGeneration,
			Table:    table,
			Priority: 4,
		}
		pg.taskQueue <- task
	}

	// Queue test generation if enabled
	if pg.config.WithTests {
		for _, table := range schema.Tables {
			task := GenerationTask{
				Type:     TestGeneration,
				Table:    table,
				Priority: 5,
			}
			pg.taskQueue <- task
		}
	}
}

// collectResults collects and logs generation results
func (pg *ParallelGenerator) collectResults() {
	successCount := 0
	errorCount := 0

	for result := range pg.results {
		if result.Success {
			successCount++
			slog.Debug("Task completed successfully",
				"table", result.Task.Table.Name,
				"type", result.Task.Type,
				"duration", result.Duration)
		} else {
			errorCount++
			slog.Error("Task failed",
				"table", result.Task.Table.Name,
				"type", result.Task.Type,
				"error", result.Error,
				"duration", result.Duration)
		}
	}

	slog.Info("Generation results",
		"successful", successCount,
		"failed", errorCount)
}

// Cleanup releases resources
func (pg *ParallelGenerator) Cleanup() {
	if pg.cancel != nil {
		pg.cancel()
	}
}

// Individual generation methods for parallel processing

// generateSingleModel generates a model for a single table
func (pg *ParallelGenerator) generateSingleModel(table introspector.Table) error {
	schema := &introspector.Schema{Tables: []introspector.Table{table}}
	return pg.Generator.generateModels(schema)
}

// generateSingleInterface generates a repository interface for a single table
func (pg *ParallelGenerator) generateSingleInterface(table introspector.Table) error {
	schema := &introspector.Schema{Tables: []introspector.Table{table}}
	return pg.Generator.generateRepositoryInterfaces(schema)
}

// generateSingleRepository generates a repository implementation for a single table
func (pg *ParallelGenerator) generateSingleRepository(table introspector.Table) error {
	schema := &introspector.Schema{Tables: []introspector.Table{table}}
	return pg.Generator.generateRepositoryImplementations(schema)
}

// generateSingleMock generates a mock for a single table
func (pg *ParallelGenerator) generateSingleMock(table introspector.Table) error {
	schema := &introspector.Schema{Tables: []introspector.Table{table}}
	return pg.Generator.generateMocks(schema)
}

// generateSingleTest generates tests for a single table
func (pg *ParallelGenerator) generateSingleTest(table introspector.Table) error {
	schema := &introspector.Schema{Tables: []introspector.Table{table}}
	return pg.Generator.generateTests(schema)
}
