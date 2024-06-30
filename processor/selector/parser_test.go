package interpreter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/shire/examples/calc/interpreter"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		input         string
		varname       string
		expectedValue int
	}{
		{"X = 5", "X", 5},
		{"X = 5 + 10", "X", 15},
		{"X = 5 - 10", "X", -5},
		{"X = 5 * 10", "X", 50},
		{"X = 10 / 5", "X", 2},
		{"X = 2 * 9 * 3 + 3 + 10 / 5 - 8 / 2", "X", 55},
		{"X = (2 + 5) * 10", "X", 70},
		{"X = (3 * 12) + 42", "X", 78},
		{"Y = (3 * X) + (X * 6)", "Y", 702},
	}

	ctx := interpreter.Context{}
	for _, testCase := range testCases {
		expr, err := interpreter.Parse(testCase.input)
		require.NoError(t, err)

		err = expr.Exec(ctx)
		require.NoError(t, err)
		require.Equal(t, testCase.expectedValue, ctx[testCase.varname])
	}
}
