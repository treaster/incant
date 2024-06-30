package interpreter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/shire/examples/calc/interpreter"
)

func TestStmt_Assign(t *testing.T) {
	testCases := []struct {
		varname        string
		exprValue      int
		expectedValue  int
		expectedString string
	}{
		{"X", 5, 5, "x = 5"},
	}

	for _, testCase := range testCases {
		expr := interpreter.Assignment(
			testCase.varname,
			interpreter.IntConstant(testCase.exprValue))

		ctx := interpreter.Context{}

		err := expr.Exec(ctx)
		require.NoError(t, err)
		require.Equal(t, testCase.expectedValue, ctx[testCase.varname])
	}
}
