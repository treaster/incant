package processor_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/treaster/incant/processor"

	"github.com/stretchr/testify/require"
)

func makeFileLoader(data map[string]string) processor.FileLoader {
	return processor.FileLoader{
		func(filename string) ([]byte, error) {
			s, hasFile := data[filename]
			if !hasFile {
				return nil, fmt.Errorf("no file named %q", filename)
			}

			return []byte(s), nil
		},
	}
}

func TestEvalContentFile(t *testing.T) {
	// Simple case
	{
		input := map[string]string{
			"file1.yaml": `
            key1: 10
            key2: "abcde"
            `,
		}

		expected := map[string]any{
			"key1": (10),
			"key2": "abcde",
		}

		actual, errs := processor.EvalContentFile(makeFileLoader(input), "file1.yaml")
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// With file: reference
	{
		input := map[string]string{
			"file1.yaml": `
            key1: 10
            key2: "file:file2.yaml"
                `,
			"file2.yaml": `
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

		actual, errs := processor.EvalContentFile(makeFileLoader(input), "file1.yaml")
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// more complicated
	{
		input := map[string]string{
			"file1.yaml": `
            key1: 10
            `,
			"file2.yaml": `
            key3:
                key4:
                - 5
                - "file:file1.yaml"
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

		actual, errs := processor.EvalContentFile(makeFileLoader(input), "file2.yaml")
		require.Equal(t, 0, len(errs))
		require.Equal(t, expected, actual)
	}

	// With circular file: reference -> error
	{
		input := map[string]string{
			"file1.yaml": `
            key1: 10
            key2: "file:file2.yaml"
            `,
			"file2.yaml": `
            "3": 20
            key4: "file:file1.yaml"
            `,
		}

		_, errs := processor.EvalContentFile(makeFileLoader(input), "file1.yaml")
		expected := []error{
			errors.New(`circular reference with "file1.yaml" (stack: [file:file1.yaml -> key2 -> * -> file:file2.yaml -> key4 -> *])`),
		}
		require.Equal(t, expected, errs)
	}

	// Int-ish looking key in the TOML spec gets converted to string.
	// Actually this doesn't work in YAML.
	/*
			{
				input := map[string]string{
					"file1.yaml": `
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

				actual, errs := processor.EvalContentFile(makeFileLoader(input), "file1.yaml")
				require.Equal(t, 0, len(errs))
				require.Equal(t, expected, actual)
			}
	*/
}
