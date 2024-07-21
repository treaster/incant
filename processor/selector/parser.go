package selector

import (
	"github.com/treaster/shire/lexer"
	"github.com/treaster/shire/parser"
)

const (
	EXPRESSION lexer.Token = "__EXPRESSION__"
)

func Parse(input string) (Expression, error) {
	parser.DebugPrint("Input is %q", input)

	// Prepare the parser
	scanner := NewScanner(input)
	state := parser.New(scanner, rules)
	final, err := state.Parse(EXPRESSION)
	if err != nil {
		return nil, err
	}
	return final.Value.(Expression), nil
}

var rules = []parser.Rule{
	{[]lexer.Token{INTEGER}, parser.ANY, EXPRESSION, IntConstant},
	{[]lexer.Token{IDENTIFIER}, parser.ANY, EXPRESSION, VarReference},
	//{[]lexer.Token{EXPRESSION}, lexer.EOF, EXPRESSION, Expression},
}
