package main

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/yuin/goldmark"
)

func main() {
	{
		const t = "FOO {{ RenderMarkdown .X }}"

		tTemp, err := template.New("t").
			Option("missingkey=error").
			Funcs(template.FuncMap{
				"RenderMarkdown": RenderMarkdown,
			}).
			Parse(t)
		if err != nil {
			panic(err.Error())
		}

		input := struct {
			X string
		}{
			"ABC *XXX* DEF",
		}

		var buf bytes.Buffer
		err = tTemp.Execute(&buf, input)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println(buf.String())
	}

	UseToml()
}

func RenderMarkdown(input string) (string, error) {
	var buf bytes.Buffer
	err := goldmark.Convert([]byte(input), &buf)
	return buf.String(), err
}

func UseToml() {
	var conf map[string]any
	_, err := toml.Decode("Age = 25", &conf)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("TOML", conf)
}
