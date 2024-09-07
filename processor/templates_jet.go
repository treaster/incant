package processor

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/CloudyKit/jet/v6"
)

type customLoader map[string][]byte

func (cl customLoader) Add(name string, contents []byte) {
	name = strings.TrimPrefix(name, "/")
	cl[name] = contents
}

func (cl customLoader) Open(name string) (io.ReadCloser, error) {
	name = strings.TrimPrefix(name, "/")
	contents, hasName := cl[name]
	if !hasName {
		return nil, fmt.Errorf("unrecognized template name %q", name)
	}
	return io.NopCloser(bytes.NewBuffer(contents)), nil
}

func (cl customLoader) Exists(name string) bool {
	name = strings.TrimPrefix(name, "/")
	_, hasName := cl[name]
	return hasName
}

func JetTemplateMgr(dataUrlRoot string) TemplateMgr {
	loader := customLoader{}
	set := jet.NewSet(
		loader,
		jet.WithSafeWriter(nil),
	).
		AddGlobal("RenderMarkdown", RenderMarkdown).
		AddGlobal("DataUrl", func(assetType string, assetPath string) string {
			fullPath := filepath.Join(dataUrlRoot, assetPath)
			return DataUrl(assetType, fullPath)
		}).
		AddGlobal("NowLocal", NowLocal).
		AddGlobal("NowUTC", NowUTC).
		AddGlobal("Sprintf", func(format string, a ...any) string {
			return fmt.Sprintf(format, a...)
		})

	return &jetTemplateMgr{
		loader,
		set,
	}
}

type jetTemplateMgr struct {
	loader customLoader
	set    *jet.Set
}

func (tm *jetTemplateMgr) ParseOne(tmplName string, tmplBody []byte) error {
	tm.loader.Add(tmplName, tmplBody)
	return nil
}

func (tm *jetTemplateMgr) Execute(tmplName string, tmplData any, output io.Writer) error {
	tmpl, err := tm.set.GetTemplate(tmplName)
	if err != nil {
		panic(fmt.Sprintf("error retrieving template %q: %s", tmplName, err.Error()))
	}

	return tmpl.Execute(output, nil, tmplData)
}
