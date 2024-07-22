package content_file_test

import (
	"errors"
	"testing"

	"github.com/treaster/ssg/processor/content_file"

	"github.com/stretchr/testify/require"
)

func TestEvalContents(t *testing.T) {
	// Simple case
	{
		input := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": "abcde",
			},
		}

		expected := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": "abcde",
			},
		}

		actual, errs := content_file.EvalContents(input)
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// With file: reference
	{
		input := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": "abcde",
			},
			"file2": content_file.Content{
				"key3": 20,
				"key4": "file:file1",
			},
		}

		expected := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": "abcde",
			},
			"file2": content_file.Content{
				"key3": 20,
				"key4": map[string]any{
					"key1": 10,
					"key2": "abcde",
				},
			},
		}

		actual, errs := content_file.EvalContents(input)
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// With file: reference that requires lookahead
	{
		input := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": "file:file2",
			},
			"file2": content_file.Content{
				"key3": 20,
				"key4": "abcde",
			},
		}

		expected := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": map[string]any{
					"key3": 20,
					"key4": "abcde",
				},
			},
			"file2": content_file.Content{
				"key3": 20,
				"key4": "abcde",
			},
		}

		actual, errs := content_file.EvalContents(input)
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// more complicated
	{
		input := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
			},
			"file2": content_file.Content{
				"key3": map[string]any{
					"key4": []any{
						5,
						"file:file1",
						"fghij",
					},
					"key5": 123,
				},
				"key6": "abcde",
			},
		}

		expected := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
			},
			"file2": content_file.Content{
				"key3": map[string]any{
					"key4": []any{
						5,
						map[string]any{
							"key1": 10,
						},
						"fghij",
					},
					"key5": 123,
				},
				"key6": "abcde",
			},
		}

		actual, errs := content_file.EvalContents(input)
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// With circular file: reference -> error
	{
		input := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": "file:file2",
			},
			"file2": content_file.Content{
				"key3": 20,
				"key4": "file:file1",
			},
		}

		_, errs := content_file.EvalContents(input)
		expected := []error{
			errors.New(`circular reference with "file1" (stack: [file:file1 -> key2 -> * -> file:file2 -> key4 -> *])`),
		}
		require.Equal(t, expected, errs)
	}

	// With non-string key in a map
	{
		input := map[string]content_file.Content{
			"file1": content_file.Content{
				"key1": 10,
				"key2": map[int]string{
					10: "abcde",
				},
			},
		}

		_, errs := content_file.EvalContents(input)
		expected := []error{
			errors.New(`non-string key in data map. (stack: [file:file1 -> key2 -> *])`),
		}
		require.Equal(t, expected, errs)
	}
}
