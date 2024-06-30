package interpreter

import "fmt"

func Assignment(varname string, expr Expression) Statement {
	return assignment{
		varname,
		expr,
	}
}

type assignment struct {
	varname string
	expr    Expression
}

func (st assignment) Exec(ctx Context) error {
	var output int
	err := st.expr.Eval(ctx, &output)
	if err != nil {
		return err
	}

	ctx[st.varname] = output
	return nil
}

func (st assignment) String() string {
	return fmt.Sprintf("%s = %s", st.varname, st.expr.String())
}
