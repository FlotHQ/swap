package swap

import (
	"testing"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "Simple variable substitution",
			template: "Hello, {{ .name }}!",
			context:  map[string]interface{}{"name": "World"},
			expected: "Hello, World!",
			wantErr:  false,
		},
		{
			name:     "Loop over object slice",
			template: "Users: {{ range .users }}{{ .name }}{{ end }}",
			context: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "Alice"},
					map[string]interface{}{"name": "Bob"},
					map[string]interface{}{"name": "Charlie"},
					map[string]interface{}{"name": "David"},
					map[string]interface{}{"name": "Eve"},
				},
			},
			expected: "Users: AliceBobCharlieDavidEve",
			wantErr:  false,
		},
	}

	engine := NewEngine()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Execute(tt.template, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(result) != tt.expected {
				t.Errorf("Execute() = %v, want %v", string(result), tt.expected)
			}
		})
	}
}
