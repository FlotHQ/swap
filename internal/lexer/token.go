package lexer

import "fmt"

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenLiteralString
	TokenLiteralNumber
	TokenSpace
	TokenLiteralBoolean
	TokenLParen
	TokenRParen
	TokenLDelim
	TokenAccessor
	TokenComma
	TokenRDelim
)

func (t TokenType) toString() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdentifier:
		return "Identifier"
	case TokenLiteralString:
		return "LiteralString"
	case TokenLiteralNumber:
		return "LiteralNumber"
	case TokenSpace:
		return "Space"
	case TokenLiteralBoolean:
		return "LiteralBoolean"
	case TokenLParen:
		return "LParen"
	case TokenRParen:
		return "RParen"
	case TokenLDelim:
		return "LDelim"
	case TokenAccessor:
		return "Accessor"
	case TokenComma:
		return "Comma"
	case TokenRDelim:
		return "RDelim"
	default:
		return "Unknown"
	}
}

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %s)", t.Type.toString(), t.Value)
}
