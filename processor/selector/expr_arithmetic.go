package interpreter

import (
	"fmt"

	"github.com/treaster/shire/lexer"
)

func Arithmetic(operator lexer.Token, expr1 Expression, expr2 Expression) Expression {
	return mathOp{
		operator,
		expr1,
		expr2,
	}
}

type mathOp struct {
	operator lexer.Token
	expr1    Expression
	expr2    Expression
}

func (exp mathOp) Eval(ctx Context, output *int) error {
	var v1 int
	err1 := exp.expr1.Eval(ctx, &v1)
	if err1 != nil {
		return err1
	}

	var v2 int
	err2 := exp.expr2.Eval(ctx, &v2)
	if err2 != nil {
		return err2
	}

	switch exp.operator {
	case PLUS:
		*output = v1 + v2
	case MINUS:
		*output = v1 - v2
	case STAR:
		*output = v1 * v2
	case SLASH:
		*output = v1 / v2
	default:
		return fmt.Errorf("unrecognized math operator %v", exp.operator)
	}

	return nil
}

func (exp mathOp) String() string {
	return fmt.Sprintf("%s %s %s", exp.expr1.String(), exp.operator, exp.expr2.String())
}
