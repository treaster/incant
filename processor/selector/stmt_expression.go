package interpreter

import "fmt"

func ExpressionStmt(expr Expression) Statement {
	return expressionStmt{
		expr,
	}
}

type expressionStmt struct {
	expr Expression
}

func (st expressionStmt) Exec(ctx Context) error {
	var output int
	err := st.expr.Eval(ctx, &output)
	if err != nil {
		return err
	}

	fmt.Printf("%d\n", output)
	return nil
}

func (st expressionStmt) String() string {
	return st.expr.String()
}
