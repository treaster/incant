package interpreter

import (
	"fmt"

	"github.com/treaster/shire/lexer"
	"github.com/treaster/shire/parser"
)

const (
	LEXPRESSION lexer.Token = "LEXPRESSION"
	REXPRESSION             = "REXPRESSION"
)

func Parse(input string) (Statement, error) {
	if parser.IsDebug() {
		fmt.Println("Input is", input)
	}

	// Prepare the parser
	scanner := NewScanner(input)
	state := parser.New(scanner, rules)
	final, err := state.Parse(STATEMENT)
	if err != nil {
		return nil, err
	}
	return final.(Statement), nil
}

var rules = []parser.Rule{
	{
		[]lexer.Token{IDENTIFIER}, parser.ANY,
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(EXPRESSION, VarReference(popItems[0].Value.(lexer.LexItem).Val))
			return true, nil
		},
	},
	{
		[]lexer.Token{LPAREN, EXPRESSION, RPAREN}, parser.ANY,
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(EXPRESSION, popItems[1].Value.(Expression))
			return true, nil
		},
	},
	{
		[]lexer.Token{EXPRESSION, IN, EXPRESSION}, parser.ANY,
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(EXPRESSION, Arithmetic(STAR, popItems[0].Value.(Expression), popItems[2].Value.(Expression)))
			return true, nil
		},
	},
	{
		[]lexer.Token{EXPRESSION, SLASH, EXPRESSION}, parser.ANY,
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(EXPRESSION, Arithmetic(SLASH, popItems[0].Value.(Expression), popItems[2].Value.(Expression)))
			return true, nil
		},
	},
	{
		[]lexer.Token{EXPRESSION}, []lexer.Token{STAR, SLASH},
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(popItems[0].Type, popItems[0].Value)
			ps.Shift()
			return true, nil
		},
	},
	{
		[]lexer.Token{EXPRESSION, PLUS, EXPRESSION}, parser.ANY,
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(EXPRESSION, Arithmetic(PLUS, popItems[0].Value.(Expression), popItems[2].Value.(Expression)))
			return true, nil
		},
	},
	{
		[]lexer.Token{EXPRESSION, MINUS, EXPRESSION}, parser.ANY,
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(EXPRESSION, Arithmetic(MINUS, popItems[0].Value.(Expression), popItems[2].Value.(Expression)))
			return true, nil
		},
	},
	{
		[]lexer.Token{IDENTIFIER, EQUAL, EXPRESSION}, []lexer.Token{lexer.EOF},
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(STATEMENT, Assignment(popItems[0].Value.(lexer.LexItem).Val, popItems[2].Value.(Expression)))
			return true, nil
		},
	},
	{
		[]lexer.Token{EXPRESSION}, []lexer.Token{lexer.EOF},
		func(ps parser.Engine, popItems []parser.ParseItem) (bool, error) {
			ps.Push(STATEMENT, ExpressionStmt(popItems[0].Value.(Expression)))
			return true, nil
		},
	},
}
