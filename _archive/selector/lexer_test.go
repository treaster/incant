package selector_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/incant/processor/selector"
	"github.com/treaster/shire/lexer"
)

func TestLexer(t *testing.T) {
	testCases := []struct {
		expr           string
		expectedTokens []lexer.LexItem
		expectError    string
	}{
		{
			"+ - / * ( ) 123 SomeIdent = +-/*()OtherIdent123=",
			[]lexer.LexItem{
				{selector.PLUS, "+"},
				{selector.MINUS, "-"},
				{selector.SLASH, "/"},
				{selector.STAR, "*"},
				{selector.LPAREN, "("},
				{selector.RPAREN, ")"},
				{selector.INTEGER, "123"},
				{selector.IDENTIFIER, "SomeIdent"},
				{selector.EQUAL, "="},

				{selector.PLUS, "+"},
				{selector.MINUS, "-"},
				{selector.SLASH, "/"},
				{selector.STAR, "*"},
				{selector.LPAREN, "("},
				{selector.RPAREN, ")"},
				{selector.IDENTIFIER, "OtherIdent"},
				{selector.INTEGER, "123"},
				{selector.EQUAL, "="},
			},
			"",
		},
	}

	for i, testCase := range testCases {
		// Verify that a series of tokens lex to themselves.
		// Consume each item in the expected tokens and ensure that its value is lexable.
		{
			for j, inputToken := range testCase.expectedTokens {
				scan := selector.NewScanner(inputToken.Val)
				item := scan.Scan()
				require.Equal(t, inputToken, item, inputToken.Val, fmt.Sprintf("Test case %d token %d", i, j))
			}
		}

		// Verify that the long string together of multiple tokens is lexed correctly.
		{
			scan := selector.NewScanner(testCase.expr)
			for j, expected := range testCase.expectedTokens {
				item := scan.Scan()
				require.Equal(t, expected, item, fmt.Sprintf("Test case %d item %d", i, j))
			}
			final := scan.Scan()
			require.Equal(t, lexer.EOF, final.Tok, fmt.Sprintf("Test case %d", i))
		}
	}
}
