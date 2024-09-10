package lexer

import (
	"reflect"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "function call with static string",
			input: "{{formatDate(\"2024-01-01T00:00:00Z\",\"2006-01-02\")}}",
			expected: []Token{
				{TokenLDelim, "{{"},
				{TokenIdentifier, "formatDate"},
				{TokenLParen, "("},
				{TokenLiteralString, "2024-01-01T00:00:00Z"},
				{TokenComma, ","},
				{TokenLiteralString, "2006-01-02"},
				{TokenRParen, ")"},
				{TokenRDelim, "}}"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "Simple text",
			input: "Hello, World!",
			expected: []Token{
				{TokenLiteralString, "Hello, World!"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "Text with variable",
			input: "Hello, {{.name}}!",
			expected: []Token{
				{TokenLiteralString, "Hello, "},
				{TokenLDelim, "{{"},
				{TokenAccessor, ".name"},
				{TokenRDelim, "}}"},
				{TokenLiteralString, "!"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "Range loop",
			input: "{{range .items}}{{.}}{{end}}",
			expected: []Token{
				{TokenLDelim, "{{"},
				{TokenIdentifier, "range"},
				{TokenSpace, " "},
				{TokenAccessor, ".items"},
				{TokenRDelim, "}}"},
				{TokenLDelim, "{{"},
				{TokenAccessor, "."},
				{TokenRDelim, "}}"},
				{TokenLDelim, "{{"},
				{TokenIdentifier, "end"},
				{TokenRDelim, "}}"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "function call",
			input: "{{upper(.name)}}",
			expected: []Token{
				{TokenLDelim, "{{"},
				{TokenIdentifier, "upper"},
				{TokenLParen, "("},
				{TokenAccessor, ".name"},
				{TokenRParen, ")"},
				{TokenRDelim, "}}"},
				{Type: TokenEOF},
			},
		},

		{
			name:  "Nested identifiers",
			input: "{{.user.name.first}}",
			expected: []Token{
				{TokenLDelim, "{{"},
				{TokenAccessor, ".user.name.first"},
				{TokenRDelim, "}}"},
				{Type: TokenEOF},
			},
		},
		{
			name: "complex invoice template",
			input: `Invoice for: {{.customer.name}}
Address: {{.customer.address}}

Items:
{{range .items}}
  - {{.name}} (Qty: {{.quantity}}, Price: ${{.price}})
{{end}}`,
			expected: []Token{
				{TokenLiteralString, "Invoice for: "},
				{TokenLDelim, "{{"},
				{TokenAccessor, ".customer.name"},
				{TokenRDelim, "}}"},
				{TokenLiteralString, "\nAddress: "},
				{TokenLDelim, "{{"},
				{TokenAccessor, ".customer.address"},
				{TokenRDelim, "}}"},
				{TokenLiteralString, "\n\nItems:\n"},
				{TokenLDelim, "{{"},
				{TokenIdentifier, "range"},
				{TokenSpace, " "},
				{TokenAccessor, ".items"},
				{TokenRDelim, "}}"},
				{TokenLiteralString, "\n  - "},
				{TokenLDelim, "{{"},
				{TokenAccessor, ".name"},
				{TokenRDelim, "}}"},
				{TokenLiteralString, " (Qty: "},
				{TokenLDelim, "{{"},
				{TokenAccessor, ".quantity"},
				{TokenRDelim, "}}"},
				{TokenLiteralString, ", Price: $"},
				{TokenLDelim, "{{"},
				{TokenAccessor, ".price"},
				{TokenRDelim, "}}"},
				{TokenLiteralString, ")\n"},
				{TokenLDelim, "{{"},
				{TokenIdentifier, "end"},
				{TokenRDelim, "}}"},
				{Type: TokenEOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens := lexer.Lex()

			if !reflect.DeepEqual(tokens, tt.expected) {
				t.Errorf("\nExpected tokens %v\nGot             %v", tt.expected, tokens)
			}
		})
	}
}

func BenchmarkLexer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lexer := NewLexer("{{range .items}}{{.}}{{end}}")
		lexer.Lex()
		lexer.Release()
	}
}
