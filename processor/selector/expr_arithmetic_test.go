package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/shire/examples/calc/interpreter"
	"github.com/treaster/shire/lexer"
)

func TestExpr_Arithmetic(t *testing.T) {
	testCases := []struct {
		op             lexer.Token
		v1             int
		v2             int
		expectedValue  int
		expectedString string
	}{
		{interpreter.PLUS, 5, 7, 12, "5 + 7"},
		{interpreter.MINUS, 5, 7, -2, "5 - 7"},
		{interpreter.STAR, 5, 7, 35, "5 * 7"},
		{interpreter.SLASH, 17, 7, 2, "17 / 7"},
	}

	for _, testCase := range testCases {
		testName := fmt.Sprintf("%v %s %v", testCase.v1, testCase.op, testCase.v2)

		expr := interpreter.Arithmetic(
			testCase.op,
			interpreter.IntConstant(testCase.v1),
			interpreter.IntConstant(testCase.v2))

		ctx := interpreter.Context{}

		var output int
		err := expr.Eval(ctx, &output)
		require.NoError(t, err, testName)
		require.Equal(t, testCase.expectedValue, output, testName)
		require.Equal(t, testCase.expectedString, expr.String())
	}
}
