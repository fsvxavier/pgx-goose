package performance

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/fsvxavier/pgx-goose/internal/interfaces"
)

// TemplateOptimizerImpl implements interfaces.TemplateOptimizer
type TemplateOptimizerImpl struct {
	cache   map[string]*CachedTemplate
	mu      sync.RWMutex
	maxSize int
	stats   *CacheStatsImpl
	funcMap template.FuncMap
}

// CachedTemplate wraps a compiled template with metadata
type CachedTemplate struct {
	template   *template.Template
	content    string
	compiledAt time.Time
	lastUsed   time.Time
	useCount   int64
}

// CacheStatsImpl implements interfaces.CacheStats
type CacheStatsImpl struct {
	mu        sync.RWMutex
	hits      int64
	misses    int64
	evictions int64
	size      int
	maxSize   int
}

// CompiledTemplateImpl implements interfaces.CompiledTemplate
type CompiledTemplateImpl struct {
	template *template.Template
	name     string
}

// NewTemplateOptimizer creates a new template optimizer with caching
func NewTemplateOptimizer(maxSize int, funcMap template.FuncMap) interfaces.TemplateOptimizer {
	if funcMap == nil {
		funcMap = getDefaultFuncMap()
	}

	return &TemplateOptimizerImpl{
		cache:   make(map[string]*CachedTemplate),
		maxSize: maxSize,
		stats: &CacheStatsImpl{
			maxSize: maxSize,
		},
		funcMap: funcMap,
	}
}

func (t *TemplateOptimizerImpl) GetTemplate(name, content string) (interfaces.CompiledTemplate, error) {
	key := t.generateKey(name, content)

	t.mu.RLock()
	cached, exists := t.cache[key]
	if exists {
		cached.lastUsed = time.Now()
		cached.useCount++
		t.mu.RUnlock()

		t.stats.recordHit()
		return &CompiledTemplateImpl{
			template: cached.template,
			name:     name,
		}, nil
	}
	t.mu.RUnlock()

	t.stats.recordMiss()

	// Compile template
	tmpl, err := template.New(name).Funcs(t.funcMap).Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to compile template %s: %w", name, err)
	}

	// Cache the compiled template
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if we need to evict
	if len(t.cache) >= t.maxSize {
		t.evictLRU()
	}

	t.cache[key] = &CachedTemplate{
		template:   tmpl,
		content:    content,
		compiledAt: time.Now(),
		lastUsed:   time.Now(),
		useCount:   1,
	}

	t.stats.size = len(t.cache)

	return &CompiledTemplateImpl{
		template: tmpl,
		name:     name,
	}, nil
}

func (t *TemplateOptimizerImpl) ExecuteTemplate(template interfaces.CompiledTemplate, data interface{}) ([]byte, error) {
	impl, ok := template.(*CompiledTemplateImpl)
	if !ok {
		return nil, fmt.Errorf("invalid template implementation")
	}

	var buf bytes.Buffer
	if err := impl.template.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", impl.name, err)
	}

	return buf.Bytes(), nil
}

func (t *TemplateOptimizerImpl) ClearCache() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cache = make(map[string]*CachedTemplate)
	t.stats.size = 0
}

func (t *TemplateOptimizerImpl) PrecompileTemplates(templates map[string]string) error {
	for name, content := range templates {
		_, err := t.GetTemplate(name, content)
		if err != nil {
			return fmt.Errorf("failed to precompile template %s: %w", name, err)
		}
	}
	return nil
}

func (t *TemplateOptimizerImpl) GetCacheStats() interfaces.CacheStats {
	t.stats.mu.RLock()
	defer t.stats.mu.RUnlock()

	return interfaces.CacheStats{
		Hits:      t.stats.hits,
		Misses:    t.stats.misses,
		Evictions: t.stats.evictions,
		Size:      t.stats.size,
		MaxSize:   t.stats.maxSize,
		HitRatio:  t.stats.GetHitRatio(),
	}
}

func (t *TemplateOptimizerImpl) generateKey(name, content string) string {
	hash := md5.Sum([]byte(name + ":" + content))
	return fmt.Sprintf("%x", hash)
}

func (t *TemplateOptimizerImpl) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range t.cache {
		if oldestKey == "" || cached.lastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.lastUsed
		}
	}

	if oldestKey != "" {
		delete(t.cache, oldestKey)
		t.stats.recordEviction()
	}
}

// CompiledTemplateImpl methods
func (c *CompiledTemplateImpl) Execute(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := c.template.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", c.name, err)
	}
	return buf.Bytes(), nil
}

func (c *CompiledTemplateImpl) Name() string {
	return c.name
}

// CacheStatsImpl methods
func (c *CacheStatsImpl) recordHit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hits++
}

func (c *CacheStatsImpl) recordMiss() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.misses++
}

func (c *CacheStatsImpl) recordEviction() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictions++
}

func (c *CacheStatsImpl) GetHits() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits
}

func (c *CacheStatsImpl) GetMisses() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.misses
}

func (c *CacheStatsImpl) GetEvictions() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.evictions
}

func (c *CacheStatsImpl) GetSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

func (c *CacheStatsImpl) GetMaxSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxSize
}

func (c *CacheStatsImpl) GetHitRatio() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0.0
	}
	return float64(c.hits) / float64(total)
}

// getDefaultFuncMap returns the default template functions
func getDefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"toPascalCase": toPascalCase,
		"toSnakeCase":  toSnakeCase,
		"lower":        func(s string) string { return s },
		"add": func(a, b int) int {
			return a + b
		},
		"slice": func(s string, start, end int) string {
			if start >= len(s) {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"join": func(sep string, elems []string) string {
			return ""
		},
		"quote": func(s string) string {
			return `"` + s + `"`
		},
		"backtick": func(s string) string {
			return "`" + s + "`"
		},
		"indent": func(spaces int, text string) string {
			return text
		},
	}
}

// Utility functions
func toPascalCase(s string) string {
	// Simple implementation - should be improved for production
	return s
}

func toSnakeCase(s string) string {
	// Simple implementation - should be improved for production
	return s
}
