package content_file_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/treaster/ssg/processor/content_file"

	"github.com/stretchr/testify/require"
)

func makeFileLoader(data map[string]string) func(string) ([]byte, error) {
	return func(filename string) ([]byte, error) {
		s, hasFile := data[filename]
		if !hasFile {
			return nil, fmt.Errorf("no file named %q", filename)
		}

		return []byte(s), nil
	}
}

func TestEvalContentFile(t *testing.T) {
	// Simple case
	{
		input := map[string]string{
			"file1": `
            key1: 10
            key2: "abcde"
            `,
		}

		expected := map[string]any{
			"key1": (10),
			"key2": "abcde",
		}

		actual, errs := content_file.EvalContentFile(makeFileLoader(input), "file1")
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// With file: reference
	{
		input := map[string]string{
			"file1": `
            key1: 10
            key2: "file:file2"
                `,
			"file2": `
                key3: 20
                key4: "abcde"
                `,
		}

		expected := map[string]any{
			"key1": (10),
			"key2": map[string]any{
				"key3": (20),
				"key4": "abcde",
			},
		}

		actual, errs := content_file.EvalContentFile(makeFileLoader(input), "file1")
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// more complicated
	{
		input := map[string]string{
			"file1": `
            key1: 10
            `,
			"file2": `
            key3:
                key4:
                - 5
                - "file:file1"
                - "fghij"
                key5: 123
            key6: "abcde"
            `,
		}

		expected := map[string]any{
			"key3": map[string]any{
				"key4": []any{
					(5),
					map[string]any{
						"key1": (10),
					},
					"fghij",
				},
				"key5": (123),
			},
			"key6": "abcde",
		}

		actual, errs := content_file.EvalContentFile(makeFileLoader(input), "file2")
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// With circular file: reference -> error
	{
		input := map[string]string{
			"file1": `
            key1: 10
            key2: "file:file2"
            `,
			"file2": `
            "3": 20
            key4: "file:file1"
            `,
		}

		_, errs := content_file.EvalContentFile(makeFileLoader(input), "file1")
		expected := []error{
			errors.New(`circular reference with "file1" (stack: [file:file1 -> key2 -> * -> file:file2 -> key4 -> *])`),
		}
		require.Equal(t, expected, errs)
	}

	// Int-ish looking key in the TOML spec gets converted to string.
	// Actually this doesn't work in YAML.
	/*
			{
				input := map[string]string{
					"file1": `
		            key1: 10
		            key2:
		                "10": "abcde"
		            `,
				}

				expected := map[string]any{
					"key1": (10),
					"key2": map[string]any{
						10: "abcde",
					},
				}

				actual, errs := content_file.EvalContentFile(makeFileLoader(input), "file1")
				require.Equal(t, 0, len(errs))
				require.Equal(t, expected, actual)
			}
	*/
}
