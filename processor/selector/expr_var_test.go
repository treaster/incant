package interpreter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/shire/examples/calc/interpreter"
)

func TestExpr_VarReference(t *testing.T) {
	testCases := []struct {
		name           string
		expr           interpreter.Expression
		ctx            interpreter.Context
		expectedValue  int
		expectedString string
	}{
		{
			"Test INT Constant",
			interpreter.VarReference("X"),
			interpreter.Context{"X": 10},
			10,
			"X",
		},
	}

	for _, testCase := range testCases {
		var output int
		err := testCase.expr.Eval(testCase.ctx, &output)
		require.NoError(t, err, testCase.name)
		require.Equal(t, testCase.expectedValue, output)
		require.Equal(t, testCase.expectedString, testCase.expr.String())
	}
}
