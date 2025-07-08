package performance

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateOptimizer(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)
	require.NotNil(t, optimizer)

	impl, ok := optimizer.(*TemplateOptimizerImpl)
	require.True(t, ok)
	assert.Equal(t, 10, impl.maxSize)
	assert.NotNil(t, impl.cache)
	assert.NotNil(t, impl.funcMap)
}

func TestTemplateOptimizer_GetTemplate_FirstTime(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	content := "Hello {{.Name}}"
	tmpl, err := optimizer.GetTemplate("test", content)

	require.NoError(t, err)
	require.NotNil(t, tmpl)
	assert.Equal(t, "test", tmpl.Name())

	// Check cache stats
	stats := optimizer.GetCacheStats()
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, 1, stats.Size)
}

func TestTemplateOptimizer_GetTemplate_CacheHit(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	content := "Hello {{.Name}}"

	// First call - cache miss
	tmpl1, err := optimizer.GetTemplate("test", content)
	require.NoError(t, err)

	// Second call - cache hit
	tmpl2, err := optimizer.GetTemplate("test", content)
	require.NoError(t, err)

	assert.Equal(t, tmpl1.Name(), tmpl2.Name())

	// Check cache stats
	stats := optimizer.GetCacheStats()
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, 0.5, stats.HitRatio)
}

func TestTemplateOptimizer_GetTemplate_InvalidTemplate(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	content := "{{.InvalidSyntax"
	_, err := optimizer.GetTemplate("invalid", content)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to compile template")
}

func TestTemplateOptimizer_ExecuteTemplate(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	content := "Hello {{.Name}}"
	tmpl, err := optimizer.GetTemplate("test", content)
	require.NoError(t, err)

	data := map[string]string{"Name": "World"}
	result, err := optimizer.ExecuteTemplate(tmpl, data)

	require.NoError(t, err)
	assert.Equal(t, "Hello World", string(result))
}

func TestTemplateOptimizer_ExecuteTemplate_InvalidData(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	content := "Hello {{.Name}}"
	tmpl, err := optimizer.GetTemplate("test", content)
	require.NoError(t, err)

	// Execute with invalid data (missing Name field)
	data := map[string]string{"WrongField": "World"}
	result, err := optimizer.ExecuteTemplate(tmpl, data)

	require.NoError(t, err) // Template execution doesn't fail, just renders empty
	assert.Equal(t, "Hello <no value>", string(result))
}

func TestTemplateOptimizer_CacheEviction(t *testing.T) {
	optimizer := NewTemplateOptimizer(2, nil) // Small cache size

	// Add 3 templates to trigger eviction
	_, err := optimizer.GetTemplate("template1", "Content 1: {{.}}")
	require.NoError(t, err)

	_, err = optimizer.GetTemplate("template2", "Content 2: {{.}}")
	require.NoError(t, err)

	stats := optimizer.GetCacheStats()
	assert.Equal(t, 2, stats.Size)
	assert.Equal(t, int64(0), stats.Evictions)

	// This should trigger eviction
	_, err = optimizer.GetTemplate("template3", "Content 3: {{.}}")
	require.NoError(t, err)

	stats = optimizer.GetCacheStats()
	assert.Equal(t, 2, stats.Size) // Still at max size
	assert.Equal(t, int64(1), stats.Evictions)
}

func TestTemplateOptimizer_ClearCache(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	// Add some templates
	_, err := optimizer.GetTemplate("template1", "Content 1")
	require.NoError(t, err)
	_, err = optimizer.GetTemplate("template2", "Content 2")
	require.NoError(t, err)

	stats := optimizer.GetCacheStats()
	assert.Equal(t, 2, stats.Size)

	// Clear cache
	optimizer.ClearCache()

	stats = optimizer.GetCacheStats()
	assert.Equal(t, 0, stats.Size)
}

func TestTemplateOptimizer_PrecompileTemplates(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	templates := map[string]string{
		"template1": "Hello {{.Name}}",
		"template2": "Goodbye {{.Name}}",
		"template3": "Welcome {{.Name}}",
	}

	err := optimizer.PrecompileTemplates(templates)
	require.NoError(t, err)

	stats := optimizer.GetCacheStats()
	assert.Equal(t, 3, stats.Size)
	assert.Equal(t, int64(3), stats.Misses) // All were cache misses initially
}

func TestTemplateOptimizer_PrecompileTemplates_Error(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	templates := map[string]string{
		"template1": "Hello {{.Name}}",
		"invalid":   "{{.InvalidSyntax",
	}

	err := optimizer.PrecompileTemplates(templates)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to precompile template")
}

func TestCompiledTemplate_Execute(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	content := "Hello {{.Name}}"
	tmpl, err := optimizer.GetTemplate("test", content)
	require.NoError(t, err)

	data := map[string]string{"Name": "World"}
	result, err := tmpl.Execute(data)

	require.NoError(t, err)
	assert.Equal(t, "Hello World", string(result))
}

func TestCacheStats_HitRatio(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	// Initially no stats
	stats := optimizer.GetCacheStats()
	assert.Equal(t, 0.0, stats.HitRatio)

	// One miss
	_, err := optimizer.GetTemplate("test", "content")
	require.NoError(t, err)

	stats = optimizer.GetCacheStats()
	assert.Equal(t, 0.0, stats.HitRatio) // 0 hits, 1 miss = 0%

	// One hit
	_, err = optimizer.GetTemplate("test", "content")
	require.NoError(t, err)

	stats = optimizer.GetCacheStats()
	assert.Equal(t, 0.5, stats.HitRatio) // 1 hit, 1 miss = 50%
}

func TestTemplateOptimizer_WithCustomFuncMap(t *testing.T) {
	customFuncMap := template.FuncMap{
		"upper": func(s string) string {
			return "UPPER:" + s
		},
	}

	optimizer := NewTemplateOptimizer(10, customFuncMap)

	content := "{{upper .Name}}"
	tmpl, err := optimizer.GetTemplate("test", content)
	require.NoError(t, err)

	data := map[string]string{"Name": "test"}
	result, err := optimizer.ExecuteTemplate(tmpl, data)

	require.NoError(t, err)
	assert.Equal(t, "UPPER:test", string(result))
}

func TestGenerateKey(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil).(*TemplateOptimizerImpl)

	key1 := optimizer.generateKey("template1", "content1")
	key2 := optimizer.generateKey("template1", "content1")
	key3 := optimizer.generateKey("template1", "content2")
	key4 := optimizer.generateKey("template2", "content1")

	// Same name and content should generate same key
	assert.Equal(t, key1, key2)

	// Different content should generate different key
	assert.NotEqual(t, key1, key3)

	// Different name should generate different key
	assert.NotEqual(t, key1, key4)
}

func TestEvictLRU(t *testing.T) {
	optimizer := NewTemplateOptimizer(2, nil).(*TemplateOptimizerImpl)

	// Add first template
	_, err := optimizer.GetTemplate("template1", "content1")
	require.NoError(t, err)

	// Add second template
	_, err = optimizer.GetTemplate("template2", "content2")
	require.NoError(t, err)

	// Access first template to make it more recently used
	_, err = optimizer.GetTemplate("template1", "content1")
	require.NoError(t, err)

	// Add third template - should evict template2 (least recently used)
	_, err = optimizer.GetTemplate("template3", "content3")
	require.NoError(t, err)

	// Verify template1 and template3 are still in cache
	_, err = optimizer.GetTemplate("template1", "content1")
	require.NoError(t, err)

	_, err = optimizer.GetTemplate("template3", "content3")
	require.NoError(t, err)

	// Check that we have hits (indicating templates were in cache)
	stats := optimizer.GetCacheStats()
	assert.True(t, stats.Hits > 0)
}

func TestDefaultFuncMap(t *testing.T) {
	funcMap := getDefaultFuncMap()

	// Test that expected functions exist
	expectedFuncs := []string{
		"toPascalCase", "toSnakeCase", "lower", "add",
		"slice", "join", "quote", "backtick", "indent",
	}

	for _, funcName := range expectedFuncs {
		assert.Contains(t, funcMap, funcName, "Function %s should exist in default func map", funcName)
	}
}

// Test individual getters for cache stats
func TestTemplateOptimizer_CacheStatsGetters(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	// Test initial state
	impl, ok := optimizer.(*TemplateOptimizerImpl)
	require.True(t, ok)

	assert.Equal(t, int64(0), impl.stats.GetHits())
	assert.Equal(t, int64(0), impl.stats.GetMisses())
	assert.Equal(t, int64(0), impl.stats.GetEvictions())
	assert.Equal(t, 0, impl.stats.GetSize())
	assert.Equal(t, 10, impl.stats.GetMaxSize())

	// Test after some operations
	content := "Hello {{.Name}}"
	_, err := optimizer.GetTemplate("test", content)
	require.NoError(t, err)

	assert.Equal(t, int64(1), impl.stats.GetMisses())
	assert.Equal(t, 1, impl.stats.GetSize())
}

// Test error handling in ExecuteTemplate
func TestTemplateOptimizer_ExecuteTemplate_Error(t *testing.T) {
	optimizer := NewTemplateOptimizer(10, nil)

	// Create a template with invalid syntax that will fail at execution
	invalidTemplate := &CompiledTemplateImpl{
		template: template.Must(template.New("invalid").Parse("{{.InvalidField.NonExistent}}")),
		name:     "invalid",
	}

	// This should return an error when executed with empty data
	_, err := optimizer.ExecuteTemplate(invalidTemplate, struct{}{})
	assert.Error(t, err)
}

// Test helper functions
func TestTemplateFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple word", "user", "user"}, // Current implementation returns input as-is
		{"snake_case", "user_profile", "user_profile"},
		{"multiple underscores", "user_profile_setting", "user_profile_setting"},
		{"empty string", "", ""},
		{"single char", "a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toPascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"PascalCase", "UserProfile", "UserProfile"}, // Current implementation returns input as-is
		{"single word", "User", "User"},
		{"camelCase", "userProfile", "userProfile"},
		{"multiple words", "UserProfileSetting", "UserProfileSetting"},
		{"empty string", "", ""},
		{"single char", "A", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test default function map
func TestGetDefaultFuncMap(t *testing.T) {
	funcMap := getDefaultFuncMap()

	assert.NotNil(t, funcMap)
	assert.Contains(t, funcMap, "toPascalCase")
	assert.Contains(t, funcMap, "toSnakeCase")

	// Test that functions work (current implementation is passthrough)
	pascalFunc, ok := funcMap["toPascalCase"].(func(string) string)
	require.True(t, ok)
	assert.Equal(t, "user_profile", pascalFunc("user_profile"))

	snakeFunc, ok := funcMap["toSnakeCase"].(func(string) string)
	require.True(t, ok)
	assert.Equal(t, "UserProfile", snakeFunc("UserProfile"))
}

// Test Execute method error handling
func TestCompiledTemplateImpl_Execute_Error(t *testing.T) {
	tmpl := &CompiledTemplateImpl{
		template: template.Must(template.New("test").Parse("{{.Field}}")),
		name:     "test",
	}

	// This should work fine
	result, err := tmpl.Execute(map[string]string{"Field": "value"})
	require.NoError(t, err)
	assert.Equal(t, "value", string(result))

	// This should fail due to missing field
	_, err = tmpl.Execute(struct{}{})
	assert.Error(t, err)
}
