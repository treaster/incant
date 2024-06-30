package interpreter_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/shire/examples/calc/interpreter"
	"github.com/treaster/shire/lexer"
)

func TestLexer(t *testing.T) {
	testCases := []struct {
		expr           string
		expectedTokens []lexer.LexItem
		expectError    string
	}{
		{
			"( ) or SomeIdent and = in not",
			[]lexer.LexItem{
				{interpreter.LPAREN, "("},
				{interpreter.RPAREN, ")"},
				{interpreter.OR, "or"},
				{interpreter.IDENTIFIER, "SomeIdent"},
				{interpreter.AND, "and"},
				{interpreter.EQUAL, "="},
				{interpreter.IN, "in"},
				{interpreter.NOT, "not"},
			},
			"",
		},
	}

	for i, testCase := range testCases {
		// Verify that a series of tokens lex to themselves.
		// Consume each item in the expected tokens and ensure that its value is lexable.
		{
			for j, inputToken := range testCase.expectedTokens {
				scan := interpreter.NewScanner(inputToken.Val)
				item := scan.Scan()
				require.Equal(t, inputToken, item, inputToken.Val, fmt.Sprintf("Test case %d token %d", i, j))
			}
		}

		// Verify that the long string together of multiple tokens is lexed correctly.
		{
			scan := interpreter.NewScanner(testCase.expr)
			for j, expected := range testCase.expectedTokens {
				item := scan.Scan()
				require.Equal(t, expected, item, fmt.Sprintf("Test case %d item %d", i, j))
			}
			final := scan.Scan()
			require.Equal(t, lexer.EOF, final.Tok, fmt.Sprintf("Test case %d", i))
		}
	}
}
