package lexer

import (
	"fmt"
	"sync"
)

type Lexer struct {
	input  string
	pos    int
	start  int
	tokens []Token
}

var lexerPool = sync.Pool{
	New: func() interface{} {
		return &Lexer{
			tokens: make([]Token, 0, 32),
		}
	},
}

func NewLexer(input string) *Lexer {
	lexer := lexerPool.Get().(*Lexer)
	lexer.input = input
	lexer.pos = 0
	lexer.start = 0
	lexer.tokens = lexer.tokens[:0]
	return lexer
}

func (l *Lexer) Release() {
	l.input = ""
	l.pos = 0
	l.start = 0
	lexerPool.Put(l)
}

func (l *Lexer) Lex() []Token {
	for l.pos < len(l.input) {
		if l.input[l.pos] == '{' && l.peek() == '{' {
			l.pos += 2
			l.addToken(TokenLDelim)
			l.lexInsideDelimiter()
		} else {
			l.lexText()
		}
	}

	if len(l.tokens) == 0 || l.tokens[len(l.tokens)-1].Type != TokenEOF {
		l.addToken(TokenEOF)
	}

	return l.tokens
}

func (l *Lexer) lexInsideDelimiter() {
	for l.pos < len(l.input) {
		if l.input[l.pos] == '}' && l.peek() == '}' {
			l.pos += 2
			l.addToken(TokenRDelim)
			return
		}

		switch {
		case isSpace(l.input[l.pos]):
			l.lexSpace()
		case l.input[l.pos] == '.':
			l.lexAccessor()
		case l.input[l.pos] == '"' || l.input[l.pos] == '\'':
			l.lexString()
		case isLetter(l.input[l.pos]):
			l.lexIdentifier()
		case isDigit(l.input[l.pos]):
			l.lexNumber()
		case l.input[l.pos] == '(':
			l.pos++
			l.addToken(TokenLParen)
		case l.input[l.pos] == ')':
			l.pos++
			l.addToken(TokenRParen)
		case l.input[l.pos] == ',':
			l.pos++
			l.addToken(TokenComma)
		default:
			panic(fmt.Sprintf("unknown token %c", l.input[l.pos]))
		}
	}
}

func (l *Lexer) lexSpace() {
	l.pos++
	for l.pos < len(l.input) && isSpace(l.input[l.pos]) {
		l.pos++
	}
	l.addToken(TokenSpace)
}

func (l *Lexer) lexString() {
	quote := l.input[l.pos]
	l.pos++
	start := l.pos
	for l.pos < len(l.input) {
		if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) && l.input[l.pos+1] == quote {
			l.pos += 2
			continue
		}
		if l.input[l.pos] == quote {
			break
		}
		l.pos++
	}
	if l.pos >= len(l.input) {
		panic("unterminated string")
	}
	l.pos++
	l.tokens = append(l.tokens, Token{Type: TokenLiteralString, Value: l.input[start : l.pos-1]})
	l.start = l.pos
}

func (l *Lexer) lexAccessor() {
	l.pos++
	for l.pos < len(l.input) && !isSpace(l.input[l.pos]) && l.input[l.pos] != '}' && l.input[l.pos] != ')' && l.input[l.pos] != '(' && l.input[l.pos] != ',' {
		l.pos++
	}
	l.addToken(TokenAccessor)
}

func (l *Lexer) lexText() {
	for l.pos < len(l.input) && (l.input[l.pos] != '{' || l.peek() != '{') {
		l.pos++
	}
	if l.pos > l.start {
		l.addToken(TokenLiteralString)
	}
}

func (l *Lexer) lexIdentifier() {
	l.pos++
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos])) {
		l.pos++
	}
	l.addToken(TokenIdentifier)
}

func (l *Lexer) lexNumber() {
	l.pos++
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}
	l.addToken(TokenLiteralNumber)
}

func (l *Lexer) addToken(tokenType TokenType) {
	l.tokens = append(l.tokens, Token{Type: tokenType, Value: l.input[l.start:l.pos]})
	l.start = l.pos
}

func (l *Lexer) peek() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

// Inline these functions for better performance
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
