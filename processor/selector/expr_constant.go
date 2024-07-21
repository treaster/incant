package selector

import (
	"strconv"

	"github.com/treaster/shire/parser"
)

func IntConstant(args []*parser.ParseItem) parser.AstNode {
	var number int
	CheckArgs(
		args,
		IntArg{&number},
	)

	return constant{
		number,
	}
}

type constant struct {
	value int
}

func (exp constant) Eval(_ Context, output *int) error {
	*output = exp.value
	return nil
}

func (exp constant) String() string {
	return strconv.Itoa(exp.value)
}
