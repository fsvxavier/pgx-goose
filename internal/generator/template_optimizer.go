package generator

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"text/template"
	"time"
)

// TemplateCache manages compiled templates with caching and optimization
type TemplateCache struct {
	cache       map[string]*CachedTemplate
	mu          sync.RWMutex
	maxSize     int
	hitCount    int64
	missCount   int64
	compileTime time.Duration
}

// CachedTemplate represents a cached compiled template
type CachedTemplate struct {
	Template    *template.Template
	Hash        string
	LastUsed    time.Time
	UseCount    int64
	CompileTime time.Duration
}

// TemplateOptimizer optimizes template compilation and execution
type TemplateOptimizer struct {
	cache       *TemplateCache
	precompiled map[string]*template.Template
	funcMap     template.FuncMap
}

// NewTemplateOptimizer creates a new template optimizer
func NewTemplateOptimizer(maxCacheSize int) *TemplateOptimizer {
	funcMap := template.FuncMap{
		"toPascalCase": toPascalCase,
		"lower":        strings.ToLower,
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
		// Additional optimized template functions
		"join":      strings.Join,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"replace":   strings.ReplaceAll,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"toUpper":   strings.ToUpper,
		"quote":     func(s string) string { return fmt.Sprintf(`"%s"`, s) },
		"backtick":  func(s string) string { return fmt.Sprintf("`%s`", s) },
		"indent": func(spaces int, text string) string {
			indent := strings.Repeat(" ", spaces)
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = indent + line
				}
			}
			return strings.Join(lines, "\n")
		},
	}

	return &TemplateOptimizer{
		cache: &TemplateCache{
			cache:   make(map[string]*CachedTemplate),
			maxSize: maxCacheSize,
		},
		precompiled: make(map[string]*template.Template),
		funcMap:     funcMap,
	}
}

// GetTemplate gets a template with caching and optimization
func (to *TemplateOptimizer) GetTemplate(name, content string) (*template.Template, error) {
	// Generate content hash for cache key
	hash := fmt.Sprintf("%x", md5.Sum([]byte(content)))
	cacheKey := fmt.Sprintf("%s_%s", name, hash)

	// Try to get from cache first
	if tmpl := to.getFromCache(cacheKey); tmpl != nil {
		return tmpl, nil
	}

	// Compile template
	start := time.Now()
	tmpl, err := template.New(name).Funcs(to.funcMap).Parse(content)
	compileTime := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("failed to compile template %s: %w", name, err)
	}

	// Store in cache
	to.storeInCache(cacheKey, tmpl, hash, compileTime)

	return tmpl, nil
}

// getFromCache retrieves a template from cache
func (to *TemplateOptimizer) getFromCache(key string) *template.Template {
	to.cache.mu.RLock()
	defer to.cache.mu.RUnlock()

	if cached, exists := to.cache.cache[key]; exists {
		cached.LastUsed = time.Now()
		cached.UseCount++
		to.cache.hitCount++

		slog.Debug("Template cache hit", "key", key, "use_count", cached.UseCount)
		return cached.Template
	}

	to.cache.missCount++
	return nil
}

// storeInCache stores a template in cache
func (to *TemplateOptimizer) storeInCache(key string, tmpl *template.Template, hash string, compileTime time.Duration) {
	to.cache.mu.Lock()
	defer to.cache.mu.Unlock()

	// Check if cache is full and evict if necessary
	if len(to.cache.cache) >= to.cache.maxSize {
		to.evictLRU()
	}

	cached := &CachedTemplate{
		Template:    tmpl,
		Hash:        hash,
		LastUsed:    time.Now(),
		UseCount:    1,
		CompileTime: compileTime,
	}

	to.cache.cache[key] = cached
	to.cache.compileTime += compileTime

	slog.Debug("Template cached", "key", key, "compile_time", compileTime)
}

// evictLRU evicts the least recently used template from cache
func (to *TemplateOptimizer) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range to.cache.cache {
		if oldestKey == "" || cached.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.LastUsed
		}
	}

	if oldestKey != "" {
		delete(to.cache.cache, oldestKey)
		slog.Debug("Template evicted from cache", "key", oldestKey)
	}
}

// PrecompileTemplates precompiles commonly used templates
func (to *TemplateOptimizer) PrecompileTemplates(templates map[string]string) error {
	slog.Info("Precompiling templates", "count", len(templates))

	for name, content := range templates {
		tmpl, err := to.GetTemplate(name, content)
		if err != nil {
			return fmt.Errorf("failed to precompile template %s: %w", name, err)
		}
		to.precompiled[name] = tmpl
	}

	return nil
}

// ExecuteTemplate executes a template with optimization
func (to *TemplateOptimizer) ExecuteTemplate(name, content string, data interface{}) (string, error) {
	tmpl, err := to.GetTemplate(name, content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// GetCacheStats returns cache statistics
func (to *TemplateOptimizer) GetCacheStats() CacheStats {
	to.cache.mu.RLock()
	defer to.cache.mu.RUnlock()

	total := to.cache.hitCount + to.cache.missCount
	hitRatio := float64(0)
	if total > 0 {
		hitRatio = float64(to.cache.hitCount) / float64(total) * 100
	}

	return CacheStats{
		Size:        len(to.cache.cache),
		MaxSize:     to.cache.maxSize,
		HitCount:    to.cache.hitCount,
		MissCount:   to.cache.missCount,
		HitRatio:    hitRatio,
		CompileTime: to.cache.compileTime,
	}
}

// CacheStats represents template cache statistics
type CacheStats struct {
	Size        int
	MaxSize     int
	HitCount    int64
	MissCount   int64
	HitRatio    float64
	CompileTime time.Duration
}

// ClearCache clears the template cache
func (to *TemplateOptimizer) ClearCache() {
	to.cache.mu.Lock()
	defer to.cache.mu.Unlock()

	to.cache.cache = make(map[string]*CachedTemplate)
	to.cache.hitCount = 0
	to.cache.missCount = 0
	to.cache.compileTime = 0

	slog.Info("Template cache cleared")
}

// WarmupCache warms up the cache with commonly used templates
func (to *TemplateOptimizer) WarmupCache(commonTemplates []string) {
	slog.Info("Warming up template cache", "templates", len(commonTemplates))

	for _, templateName := range commonTemplates {
		// This would typically load template content and compile it
		// Implementation depends on how templates are stored
		slog.Debug("Warming up template", "name", templateName)
	}
}
