package selector

import (
	"github.com/treaster/shire/lexer"
	"github.com/treaster/shire/parser"
)

const (
	PLUS       lexer.Token = "+"
	MINUS                  = "-"
	SLASH                  = "/"
	STAR                   = "*"
	LPAREN                 = "("
	RPAREN                 = ")"
	EQUAL                  = "="
	INTEGER                = "INTEGER"
	IDENTIFIER             = "IDENTIFIER"
)

// NewScanner creates a new scanner for the input string.
func NewScanner(input string) parser.Scanner {
	return lexer.New(input, lexBasic)
}

func lexBasic(l lexer.Engine) lexer.StateFn {
	symbols := map[rune]lexer.Token{
		'+': PLUS,
		'-': MINUS,
		'*': STAR,
		'/': SLASH,
		'(': LPAREN,
		')': RPAREN,
		'=': EQUAL,
	}

	r := l.Next()
	symbolTok, hasSymbol := symbols[r]
	switch {
	case r == lexer.EofRune:
		l.Emit(lexer.EOF)
		return nil
	case hasSymbol:
		l.Emit(symbolTok)
		return lexBasic
	case lexer.IsWhitespace(r):
		l.Ignore()
	case lexer.IsLetter(r):
		_ = l.AcceptRun(lexer.IsLetter)
		l.Emit(IDENTIFIER)
		return lexBasic
	case lexer.IsDigit(r):
		l.Backup()
		return lexInteger
	}
	return lexBasic
}

func lexInteger(l lexer.Engine) lexer.StateFn {
	l.AcceptRun(lexer.IsDigit)
	l.Emit(INTEGER)
	return lexBasic
}
