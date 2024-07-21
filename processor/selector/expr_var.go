package selector

import (
	"fmt"

	"github.com/treaster/shire/parser"
)

func VarReference(args []*parser.ParseItem) parser.AstNode {
	var identifier string
	CheckArgs(
		args,
		IdentifierArg{&identifier},
	)

	return varref{
		identifier,
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
