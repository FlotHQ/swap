package swap

import (
	"testing"
)

func generateTemplate() string {
	return "Hello, {{ .name }}! %s\n"
}

func BenchmarkExecute(b *testing.B) {
	engine := NewEngine()
	context := map[string]interface{}{"name": "World"}
	template := generateTemplate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Execute(template, context)
	}
}

func BenchmarkExecuteWithCache(b *testing.B) {
	engine := NewEngine(WithCacheEnabled(true))
	context := map[string]interface{}{"name": "World"}
	template := generateTemplate()
	b.ResetTimer()
	_, _ = engine.Execute(template, context)
	for i := 0; i < b.N; i++ {
		_, _ = engine.Execute(template, context)
	}
}

func BenchmarkRun(b *testing.B) {
	engine := NewEngine()
	template := generateTemplate()
	program, err := engine.Compile(template)
	if err != nil {
		b.Fatal(err)
	}
	context := map[string]interface{}{"name": "World"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Run(program, context)
	}
}
