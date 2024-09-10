package compiler

import (
	"testing"

	"github.com/flothq/swap/internal/lexer"
	"github.com/flothq/swap/pkg/bytecode"
)

func TestCompiler(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []lexer.Token
		expected []bytecode.Instruction
		wantErr  bool
	}{
		{
			name: "Simple text",
			tokens: []lexer.Token{
				{lexer.TokenLiteralString, "Hello, World!"},
				{Type: lexer.TokenEOF},
			},
			expected: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
		},
		{
			name: "Text with variable",
			tokens: []lexer.Token{
				{Type: lexer.TokenLiteralString, Value: "Hello, "},
				{Type: lexer.TokenLDelim},
				{Type: lexer.TokenIdentifier, Value: ".name"},
				{Type: lexer.TokenRDelim},
				{Type: lexer.TokenLiteralString, Value: "!"},
				{Type: lexer.TokenEOF},
			},
			expected: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
				bytecode.PackInstruction(bytecode.OpPrintConst, 2, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
		},
		{
			name: "Range loop",
			tokens: []lexer.Token{

				{Type: lexer.TokenLiteralString, Value: "Item: "},
				{Type: lexer.TokenLDelim},
				{Type: lexer.TokenIdentifier, Value: "range"},
				{Type: lexer.TokenSpace},
				{Type: lexer.TokenAccessor, Value: ".items"},
				{Type: lexer.TokenRDelim},
				{Type: lexer.TokenLiteralString, Value: "Item: "},
				{Type: lexer.TokenLDelim},
				{Type: lexer.TokenAccessor, Value: "."},
				{Type: lexer.TokenRDelim},
				{Type: lexer.TokenLDelim},
				{Type: lexer.TokenIdentifier, Value: "end"},
				{Type: lexer.TokenRDelim},
				{Type: lexer.TokenEOF},
			},
			expected: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpLoopStart, 1, 0, 0),
				bytecode.PackInstruction(bytecode.OpPrintConst, 2, 0, 0),
				bytecode.PackInstruction(bytecode.OpResolvePrint, 3, 0, 0),
				bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
		},
		{
			name: "Nested identifiers",
			tokens: []lexer.Token{
				{Type: lexer.TokenLDelim},
				{Type: lexer.TokenIdentifier, Value: ".user.name.first"},
				{Type: lexer.TokenRDelim},
				{Type: lexer.TokenEOF},
			},
			expected: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpResolvePrint, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
		},
		{
			name: "function call",
			tokens: []lexer.Token{
				{Type: lexer.TokenLDelim, Value: "{{"},
				{Type: lexer.TokenIdentifier, Value: "upper"},
				{Type: lexer.TokenLParen, Value: "("},
				{Type: lexer.TokenAccessor, Value: ".name"},
				{Type: lexer.TokenRParen, Value: ")"},
				{Type: lexer.TokenRDelim, Value: "}}"},
				{Type: lexer.TokenEOF},
			},
			expected: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpResolveLoad, 0, 1, 0), // Updated this line
				bytecode.PackInstruction(bytecode.OpCall, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(tt.tokens)
			instructions, _, err := compiler.Compile(tt.tokens)

			if (err != nil) != tt.wantErr {
				t.Errorf("Compiler.Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(instructions) != len(tt.expected) {
					t.Errorf("Instruction count mismatch. Expected %d, got %d", len(tt.expected), len(instructions))
					return
				}

				for i, exp := range tt.expected {
					got := instructions[i]
					var gotUnpacked, expUnpacked bytecode.UnpackedInstruction
					gotUnpacked.Unpack(got)
					expUnpacked.Unpack(exp)

					if gotUnpacked.Op != expUnpacked.Op {
						t.Errorf("Instruction %d: Opcode mismatch. Expected %v, got %v", i, expUnpacked.Op, gotUnpacked.Op)
					}
					if gotUnpacked.A != expUnpacked.A || gotUnpacked.B != expUnpacked.B || gotUnpacked.C != expUnpacked.C {
						t.Errorf("Instruction %d: Operands mismatch. Expected (%v, %v, %v), got (%v, %v, %v)",
							i, expUnpacked.A, expUnpacked.B, expUnpacked.C, gotUnpacked.A, gotUnpacked.B, gotUnpacked.C)
					}
				}
			}
		})
	}
}
