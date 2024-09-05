package processor

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func RenderMarkdown(input string) (string, error) {
	md := goldmark.New(
		// Enables table, strikethrough, linkify, and tasklist markdown features.
		goldmark.WithExtensions(extension.GFM),
	)

	var buf bytes.Buffer
	err := md.Convert([]byte(input), &buf)
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

var timeLayouts = map[string]string{
	"Layout":      time.Layout,
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
	"DateTime":    time.DateTime,
	"DateOnly":    time.DateOnly,
	"TimeOnly":    time.TimeOnly,
}

func NowLocal(localeName string, layout string) string {
	location, err := time.LoadLocation(localeName)
	if err != nil {
		return fmt.Sprintf("error in NowLocal(): %s", err.Error())
	}

	finalLayout, hasLayout := timeLayouts[layout]
	if !hasLayout {
		finalLayout = layout
	}

	return time.Now().In(location).Format(finalLayout)
}

func NowUTC(layout string) string {
	finalLayout, hasLayout := timeLayouts[layout]
	if !hasLayout {
		finalLayout = layout
	}

	return time.Now().Format(finalLayout)
}

func NamedArgs(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}
