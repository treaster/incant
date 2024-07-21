package selector_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster.net/ssg/processor/selector"
)

type ContextVar struct {
	Int    int
	Map    map[string]ContextVar
	Array  []ContextVar
	Nested *ContextVar
}

func TestParser(t *testing.T) {
	testFunc := func(
		input string,
		ctx map[string]ContextVar,
		expectedValue int,
	) {
		expr, err := selector.Parse(input)
		require.NoError(t, err)

		var result int
		err = expr.Eval(ctx, &result)
		require.NoError(t, err)
		require.Equal(t, expectedValue, result, fmt.Sprintf("Input is %q", input))
	}

	testFunc("5", map[string]ContextVar{}, 5)
	// {"Context.Int", Context{Int: 5}, 5},

}
