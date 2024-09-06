package processor

import (
	"fmt"
	"io"
	"path/filepath"
	"text/template"
)

func GoTemplateMgr(dataUrlRoot string) TemplateMgr {
	tmpl := template.
		New("incant").
		Funcs(template.FuncMap{
			"RenderMarkdown": RenderMarkdown,
			"DataUrl": func(assetType string, assetPath string) string {
				fullPath := filepath.Join(dataUrlRoot, assetPath)
				return DataUrl(assetType, fullPath)
			},
			"Add":       func(a int, b int) int { return a + b },
			"Sub":       func(a int, b int) int { return a - b },
			"Mult":      func(a int, b int) int { return a * b },
			"Div":       func(a int, b int) int { return a / b },
			"NowLocal":  NowLocal,
			"NowUTC":    NowUTC,
			"NamedArgs": NamedArgs,
		}).
		Option("missingkey=error")

	return &goTemplateMgr{tmpl}
}

type goTemplateMgr struct {
	tmpl *template.Template
}

func (tm *goTemplateMgr) ParseOne(tmplName string, tmplBody []byte) error {
	_, err := tm.tmpl.New(tmplName).Parse(string(tmplBody))
	if err != nil {
		return fmt.Errorf("error parsing template %q: %s", tmplName, err.Error())
	}
	return nil
}

func (tm *goTemplateMgr) Execute(tmplName string, tmplData any, output io.Writer) error {
	tmpl := tm.tmpl.Lookup(tmplName)
	if tmpl == nil {
		panic(fmt.Sprintf("error: template %q not found", tmplName))
	}

	return tmpl.Execute(output, tmplData)
}
