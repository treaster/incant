package interpreter

import "fmt"

func VarReference(varname string) Expression {
	return varref{
		varname,
	}
}

type varref struct {
	varname string
}

func (exp varref) Eval(ctx Context, output *int) error {
	value, hasVar := ctx[exp.varname]
	if !hasVar {
		return fmt.Errorf("unrecognized variable name %q", exp.varname)
	}
	*output = value
	return nil
}

func (exp varref) String() string {
	return exp.varname
}
