package processor

import (
	"bytes"

	"github.com/yuin/goldmark"
)

func RenderMarkdown(input string) (string, error) {
	var buf bytes.Buffer
	err := goldmark.Convert([]byte(input), &buf)
	return buf.String(), err
}
