package selector

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/treaster/shire/lexer"
	"github.com/treaster/shire/parser"
)

type Checker interface {
	Check(*parser.ParseItem) error
}

func CheckArgs(
	args []*parser.ParseItem,
	checkers ...Checker,
) {
	if len(args) != len(checkers) {
		panic("wrong number of tokens")
	}

	for i, arg := range args {
		checker := checkers[i]
		err := checker.Check(arg)
		if err != nil {
			panic(err.Error())
		}
	}
}

type ExprArg struct {
	expr *Expression
}

func (c ExprArg) Check(parseItem *parser.ParseItem) error {
	if parseItem.Type != EXPRESSION {
		return errors.New("wrong type")
	}

	expr, isExpr := parseItem.Value.(Expression)
	if !isExpr {
		return errors.New("not really an expression")
	}

	*c.expr = expr
	return nil
}

type TokenArg struct {
	token   *lexer.Token
	allowed []lexer.Token
}

func (c TokenArg) Check(parseItem *parser.ParseItem) error {
	found := false
	for _, tok := range c.allowed {
		found = found || tok == parseItem.Type
	}

	if !found {
		return errors.New("did not find expected token type")
	}

	if c.token != nil {
		*c.token = parseItem.Type
	}
	return nil
}

type IntArg struct {
	number *int
}

func (c IntArg) Check(parseItem *parser.ParseItem) error {
	if parseItem.Type != INTEGER {
		return errors.New("wrong type")
	}

	strVal, isStr := parseItem.Value.(string)
	if !isStr {
		panic(fmt.Sprintf("doesn't look like a lexer token %v, %q", parseItem.Value, strVal))
	}

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		panic(fmt.Sprintf("int didn't parse: %s", err.Error()))
	}

	*c.number = intVal
	return nil
}

type IdentifierArg struct {
	identifier *string
}

func (c IdentifierArg) Check(parseItem *parser.ParseItem) error {
	if parseItem.Type != IDENTIFIER {
		panic("wrong token type")
	}

	strVal, isStr := parseItem.Value.(string)
	if !isStr {
		panic(fmt.Sprintf("doesn't look like an identifier: %q", parseItem.Value))
	}

	*c.identifier = strVal
	return nil
}
