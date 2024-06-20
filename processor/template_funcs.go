package processor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
)

func RenderMarkdown(input string) (string, error) {
	var buf bytes.Buffer
	err := goldmark.Convert([]byte(input), &buf)
	return buf.String(), err
}

func DataUrl(assetType string, assetPath string) string {
	data, err := os.ReadFile(assetPath)
	if err != nil {
		panic(fmt.Sprintf("error reading source asset %s: %s", assetPath, err.Error()))
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", assetType, encoded)
}
