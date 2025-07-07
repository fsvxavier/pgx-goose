package generator

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateOptimizer(t *testing.T) {
	optimizer := NewTemplateOptimizer(10)

	assert.NotNil(t, optimizer)
	assert.NotNil(t, optimizer.cache)
	assert.NotNil(t, optimizer.funcMap)
	assert.Equal(t, 10, optimizer.cache.maxSize)
	assert.Len(t, optimizer.funcMap, 15) // Check that all template functions are added
}

func TestTemplateOptimizer_GetTemplate(t *testing.T) {
	optimizer := NewTemplateOptimizer(5)

	templateContent := `Hello {{.Name}}!`
	templateName := "greeting"

	t.Run("first compilation", func(t *testing.T) {
		tmpl, err := optimizer.GetTemplate(templateName, templateContent)

		require.NoError(t, err)
		assert.NotNil(t, tmpl)

		// Verify template works
		var buf strings.Builder
		err = tmpl.Execute(&buf, map[string]string{"Name": "World"})
		require.NoError(t, err)
		assert.Equal(t, "Hello World!", buf.String())

		// Check cache stats
		stats := optimizer.GetCacheStats()
		assert.Equal(t, int64(0), stats.HitCount) // First time, no hit
		assert.Equal(t, int64(1), stats.MissCount)
	})

	t.Run("cache hit", func(t *testing.T) {
		tmpl, err := optimizer.GetTemplate(templateName, templateContent)

		require.NoError(t, err)
		assert.NotNil(t, tmpl)

		// Check cache stats
		stats := optimizer.GetCacheStats()
		assert.Equal(t, int64(1), stats.HitCount)
		assert.Equal(t, int64(1), stats.MissCount)
		assert.Equal(t, 50.0, stats.HitRatio)
	})

	t.Run("different content same name", func(t *testing.T) {
		differentContent := `Goodbye {{.Name}}!`
		tmpl, err := optimizer.GetTemplate(templateName, differentContent)

		require.NoError(t, err)
		assert.NotNil(t, tmpl)

		// Should be a cache miss because content hash is different
		stats := optimizer.GetCacheStats()
		assert.Equal(t, int64(1), stats.HitCount)
		assert.Equal(t, int64(2), stats.MissCount)
	})

	t.Run("invalid template", func(t *testing.T) {
		invalidContent := `Hello {{.Name`
		_, err := optimizer.GetTemplate("invalid", invalidContent)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compile template")
	})
}

func TestTemplateOptimizer_CacheEviction(t *testing.T) {
	optimizer := NewTemplateOptimizer(2) // Small cache for testing eviction

	templates := []struct {
		name    string
		content string
	}{
		{"tmpl1", "Template 1: {{.Value}}"},
		{"tmpl2", "Template 2: {{.Value}}"},
		{"tmpl3", "Template 3: {{.Value}}"}, // This should trigger eviction
	}

	// Fill cache
	for _, tmpl := range templates {
		_, err := optimizer.GetTemplate(tmpl.name, tmpl.content)
		require.NoError(t, err)
	}

	stats := optimizer.GetCacheStats()
	assert.Equal(t, 2, stats.Size) // Cache should be at max size
	assert.Equal(t, int64(0), stats.HitCount)
	assert.Equal(t, int64(3), stats.MissCount)

	// Access first template again - should be a miss if evicted
	_, err := optimizer.GetTemplate(templates[0].name, templates[0].content)
	require.NoError(t, err)

	stats = optimizer.GetCacheStats()
	// Should still be miss count 4 if first template was evicted
	assert.Equal(t, int64(4), stats.MissCount)
}

func TestTemplateOptimizer_ExecuteTemplate(t *testing.T) {
	optimizer := NewTemplateOptimizer(5)

	templateContent := `Name: {{.Name}}, Age: {{.Age}}`
	data := map[string]interface{}{
		"Name": "John",
		"Age":  30,
	}

	result, err := optimizer.ExecuteTemplate("person", templateContent, data)

	require.NoError(t, err)
	assert.Equal(t, "Name: John, Age: 30", result)
}

func TestTemplateOptimizer_TemplateFunctions(t *testing.T) {
	optimizer := NewTemplateOptimizer(5)

	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "toPascalCase",
			template: `{{toPascalCase .Value}}`,
			data:     map[string]string{"Value": "hello_world"},
			expected: "HelloWorld",
		},
		{
			name:     "lower",
			template: `{{lower .Value}}`,
			data:     map[string]string{"Value": "HELLO"},
			expected: "hello",
		},
		{
			name:     "add",
			template: `{{add .A .B}}`,
			data:     map[string]int{"A": 5, "B": 3},
			expected: "8",
		},
		{
			name:     "slice",
			template: `{{slice .Value 1 4}}`,
			data:     map[string]string{"Value": "hello"},
			expected: "ell",
		},
		{
			name:     "join",
			template: `{{join .Values ", "}}`,
			data:     map[string][]string{"Values": {"a", "b", "c"}},
			expected: "a, b, c",
		},
		{
			name:     "quote",
			template: `{{quote .Value}}`,
			data:     map[string]string{"Value": "hello"},
			expected: `"hello"`,
		},
		{
			name:     "backtick",
			template: `{{backtick .Value}}`,
			data:     map[string]string{"Value": "hello"},
			expected: "`hello`",
		},
		{
			name:     "indent",
			template: `{{indent 2 .Value}}`,
			data:     map[string]string{"Value": "line1\nline2"},
			expected: "  line1\n  line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := optimizer.ExecuteTemplate(tt.name, tt.template, tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateOptimizer_ClearCache(t *testing.T) {
	optimizer := NewTemplateOptimizer(5)

	// Add some templates to cache
	_, err := optimizer.GetTemplate("test1", "Template 1")
	require.NoError(t, err)
	_, err = optimizer.GetTemplate("test2", "Template 2")
	require.NoError(t, err)

	stats := optimizer.GetCacheStats()
	assert.Equal(t, 2, stats.Size)
	assert.Greater(t, stats.CompileTime, time.Duration(0))

	// Clear cache
	optimizer.ClearCache()

	stats = optimizer.GetCacheStats()
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, int64(0), stats.HitCount)
	assert.Equal(t, int64(0), stats.MissCount)
	assert.Equal(t, time.Duration(0), stats.CompileTime)
}

func TestTemplateOptimizer_PrecompileTemplates(t *testing.T) {
	optimizer := NewTemplateOptimizer(5)

	templates := map[string]string{
		"model":      "type {{.Name}} struct { {{range .Fields}} {{.Name}} {{.Type}} {{end}} }",
		"interface":  "type {{.Name}} interface { {{range .Methods}} {{.Name}}() {{.Return}} {{end}} }",
		"repository": "func (r *{{.Name}}) Get(id int) (*{{.Model}}, error) { return nil, nil }",
	}

	err := optimizer.PrecompileTemplates(templates)
	require.NoError(t, err)

	// Verify templates are in cache
	stats := optimizer.GetCacheStats()
	assert.Equal(t, 3, stats.Size)

	// Verify precompiled templates work
	for name, content := range templates {
		tmpl, err := optimizer.GetTemplate(name, content)
		require.NoError(t, err)
		assert.NotNil(t, tmpl)
	}

	// Should have high hit ratio since templates were precompiled
	stats = optimizer.GetCacheStats()
	assert.Equal(t, int64(3), stats.HitCount)
}

// Benchmarks

func BenchmarkTemplateOptimizer_GetTemplate_CacheHit(b *testing.B) {
	optimizer := NewTemplateOptimizer(100)
	templateContent := `Hello {{.Name}}! You are {{.Age}} years old.`

	// Pre-warm cache
	_, err := optimizer.GetTemplate("greeting", templateContent)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := optimizer.GetTemplate("greeting", templateContent)
		require.NoError(b, err)
	}
}

func BenchmarkTemplateOptimizer_GetTemplate_CacheMiss(b *testing.B) {
	optimizer := NewTemplateOptimizer(100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		templateContent := fmt.Sprintf(`Hello {{.Name}}! Template %d`, i)
		_, err := optimizer.GetTemplate(fmt.Sprintf("greeting_%d", i), templateContent)
		require.NoError(b, err)
	}
}

func BenchmarkTemplateOptimizer_ExecuteTemplate(b *testing.B) {
	optimizer := NewTemplateOptimizer(10)
	templateContent := `
Name: {{.Name}}
Age: {{.Age}}
Email: {{.Email}}
Active: {{if .Active}}Yes{{else}}No{{end}}
Tags: {{join .Tags ", "}}
`

	data := map[string]interface{}{
		"Name":   "John Doe",
		"Age":    30,
		"Email":  "john@example.com",
		"Active": true,
		"Tags":   []string{"developer", "golang", "postgres"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := optimizer.ExecuteTemplate("profile", templateContent, data)
		require.NoError(b, err)
	}
}

func BenchmarkTemplateOptimizer_CacheSize(b *testing.B) {
	cacheSizes := []int{10, 50, 100, 500}

	for _, size := range cacheSizes {
		b.Run(fmt.Sprintf("cache_size_%d", size), func(b *testing.B) {
			optimizer := NewTemplateOptimizer(size)
			templateContent := `Hello {{.Name}}!`

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Mix of cache hits and misses
				templateName := fmt.Sprintf("greeting_%d", i%size)
				_, err := optimizer.GetTemplate(templateName, templateContent)
				require.NoError(b, err)
			}
		})
	}
}
