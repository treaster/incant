package interpreter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/shire/examples/calc/interpreter"
)

func TestExpr_Constant(t *testing.T) {
	testCases := []struct {
		name           string
		expr           interpreter.Expression
		expectedValue  int
		expectedString string
	}{
		{
			"Test INT Constant",
			interpreter.IntConstant(5),
			5,
			"5",
		},
	}

	for _, testCase := range testCases {
		ctx := interpreter.Context{}
		var output int
		err := testCase.expr.Eval(ctx, &output)
		require.NoError(t, err, testCase.name)
		require.Equal(t, testCase.expectedValue, output)
		require.Equal(t, testCase.expectedString, testCase.expr.String())
	}
}
