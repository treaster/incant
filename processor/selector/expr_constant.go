package interpreter

import (
	"strconv"
)

func IntConstant(val int) Expression {
	return constant{
		val,
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
