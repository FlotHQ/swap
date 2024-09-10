package compiler

import (
	"fmt"
	"sync"

	"github.com/flothq/swap/internal/lexer"
	"github.com/flothq/swap/pkg/bytecode"
)

type Compiler struct {
	tokens       []lexer.Token
	instructions []bytecode.Instruction
	constants    []bytecode.Constant
	pos          int
}

var compilerPool = sync.Pool{
	New: func() interface{} {
		return &Compiler{
			tokens:       make([]lexer.Token, 0),
			instructions: make([]bytecode.Instruction, 0),
			constants:    make([]bytecode.Constant, 0),
			pos:          0,
		}
	},
}

func NewCompiler(tokens []lexer.Token) *Compiler {
	c := compilerPool.Get().(*Compiler)
	c.instructions = c.instructions[:0]
	c.constants = c.constants[:0]
	c.pos = 0
	c.tokens = tokens
	return c
}

func (c *Compiler) Release() {
	c.tokens = c.tokens[:0]
	compilerPool.Put(c)
}

func (c *Compiler) Compile(tokens []lexer.Token) ([]bytecode.Instruction, []bytecode.Constant, error) {
	for {
		token := c.tokens[c.pos]
		switch token.Type {
		case lexer.TokenLiteralString:
			c.constants = append(c.constants, bytecode.Constant{Type: bytecode.ConstString, Value: token.Value})
			c.emit(bytecode.OpPrintConst, uint8(len(c.constants)-1), 0, 0)
			c.pos++
		case lexer.TokenLDelim:
			if err := c.compileExpression(); err != nil {
				return nil, nil, err
			}
		case lexer.TokenEOF:
			c.emit(bytecode.OpHalt, 0, 0, 0)
			c.pos++
			return c.instructions, c.constants, nil
		default:
			return nil, nil, fmt.Errorf("invalid token in compile: %v, %s", token.Type, token.Value)
		}

		if c.pos >= len(c.tokens) {
			break
		}
	}

	return c.instructions, c.constants, nil
}

func (c *Compiler) eatWhitespace() {
	for c.pos < len(c.tokens) && c.tokens[c.pos].Type == lexer.TokenSpace {
		c.pos++
	}
}

func (c *Compiler) compileExpression() error {
	c.pos++
	c.eatWhitespace()
	for i := c.pos; i < len(c.tokens); i++ {
		c.eatWhitespace()
		token := c.tokens[c.pos]
		switch token.Type {
		case lexer.TokenRDelim:
			c.pos++
			return nil
		case lexer.TokenAccessor:
			c.constants = append(c.constants, bytecode.Constant{Type: bytecode.ConstString, Value: token.Value})
			c.emit(bytecode.OpResolvePrint, uint8(len(c.constants)-1), 0, 0)
			c.pos++
			c.eatWhitespace()
		case lexer.TokenIdentifier:
			if token.Value == "end" {
				c.pos++
				c.emit(bytecode.OpLoopEnd, 0, 0, 0)
				c.eatWhitespace()
			} else if token.Value == "range" {
				c.pos++
				if c.pos >= len(c.tokens) {

					return fmt.Errorf("unexpected end of input after 'range'")
				}
				c.pos++
				nextToken := c.tokens[c.pos]
				if nextToken.Type != lexer.TokenAccessor {
					fmt.Println(nextToken)
					return fmt.Errorf("expected identifier after 'range', got %v", nextToken)
				}
				c.pos++
				c.eatWhitespace()

				if c.pos >= len(c.tokens) || c.tokens[c.pos].Type != lexer.TokenRDelim {
					return fmt.Errorf(
						fmt.Sprintf("expected right delimiter after 'range', got %v", c.tokens[c.pos]),
					)
				}
				c.constants = append(c.constants, bytecode.Constant{Type: bytecode.ConstString, Value: nextToken.Value})
				c.emit(bytecode.OpLoopStart, uint8(len(c.constants)-1), 0, 0)
			} else {
				if c.pos+1 < len(c.tokens) && c.tokens[c.pos+1].Type == lexer.TokenLParen {
					tokens, err := c.compileFunctionCall()
					if err != nil {
						return err
					}
					for i := len(tokens) - 1; i >= 0; i-- {
						c.instructions = append(c.instructions, tokens[i])
					}
				} else {
					c.constants = append(c.constants, bytecode.Constant{Type: bytecode.ConstString, Value: token.Value})
					c.emit(bytecode.OpResolvePrint, uint8(len(c.constants)-1), 0, 0)
					c.pos++
				}
			}
		default:
			return fmt.Errorf("unexpected token in expression: %v", token)
		}
	}

	return fmt.Errorf("unexpected end of input")
}

func (c *Compiler) compileFunctionCall() ([]bytecode.Instruction, error) {
	stack := make([]bytecode.Instruction, 0)
	count := uint8(0)

	token := c.tokens[c.pos]

	if token.Type != lexer.TokenIdentifier {

		return nil, fmt.Errorf("expected identifier after 'call', got %v", token)
	}

	if c.tokens[c.pos+1].Type != lexer.TokenLParen {
		return nil, fmt.Errorf("expected '(' after function name, got %v", c.tokens[c.pos+1])
	}

	c.constants = append(c.constants, bytecode.Constant{Type: bytecode.ConstString, Value: token.Value})
	stack = append(stack, bytecode.PackInstruction(bytecode.OpCall, uint8(len(c.constants)-1), 0, 0))

	c.pos++
	c.pos++

	for {
		token := c.tokens[c.pos]
		switch token.Type {
		case lexer.TokenAccessor:
			c.constants = append(c.constants, bytecode.Constant{Type: bytecode.ConstString, Value: token.Value})
			stack = append(stack, bytecode.PackInstruction(bytecode.OpResolveLoad, count, uint8(len(c.constants)-1), 0))
			count++
		case lexer.TokenIdentifier:
			tokens, err := c.compileFunctionCall()
			if err != nil {
				return nil, err
			}
			stack = append(stack, tokens...)
		case lexer.TokenRParen:
			c.pos++
			return stack, nil
		case lexer.TokenLiteralString:
			c.constants = append(c.constants, bytecode.Constant{Type: bytecode.ConstString, Value: token.Value})
			stack = append(stack, bytecode.PackInstruction(bytecode.OpLoadConst, count, uint8(len(c.constants)-1), 0))
			count++
		case lexer.TokenSpace:
		case lexer.TokenComma:
		default:
			return nil, fmt.Errorf("unexpected token in function call: %v", token)
		}

		c.pos++
	}
}

func (c *Compiler) emit(op bytecode.OpCode, a, b, d uint8) {
	c.instructions = append(c.instructions, bytecode.PackInstruction(op, a, b, d))
}
