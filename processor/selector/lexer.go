package interpreter

import (
	"github.com/treaster/shire/lexer"
	"github.com/treaster/shire/parser"
)

const (
	LPAREN     lexer.Token = "("
	RPAREN                 = ")"
	EQUAL                  = "="
	IN                     = "IN"
	OR                     = "OR"
	AND                    = "AND"
	NOT                    = "NOT"
	IDENTIFIER             = "IDENTIFIER"
)

// NewScanner creates a new scanner for the input string.
func NewScanner(input string) parser.Scanner {
	return lexer.New(input, lexBasic)
}

func lexBasic(l lexer.Engine) lexer.StateFn {
	symbols := map[rune]lexer.Token{
		'(': LPAREN,
		')': RPAREN,
		'=': EQUAL,
	}

	keywords := map[string]lexer.Token{
		"and": AND,
		"or":  OR,
		"not": NOT,
		"in":  IN,
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
		run := l.AcceptRun(lexer.IsLetter)
		token, isKeyword := keywords[run]
		if isKeyword {
			l.Emit(token)
		} else {
			l.Emit(IDENTIFIER)
		}
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
