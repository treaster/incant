package processor_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/ssg/processor"
)

func TestEvalContents(t *testing.T) {
	input := map[string]Content{
		"file1": Content{
			"key1": 10,
			"key2": "abcde",
		},
	}

	expected := map[string]Content{
		"file1": Content{
			"key1": 10,
			"key2": "abcde",
		},
	}

	actual := processor.EvalContents(input)
	require.Equal(t, expected, actual)
}
