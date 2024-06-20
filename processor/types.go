package processor

import "text/template"

type Config struct {
	StaticRoot    string
	TemplatesRoot string
	ContentRoot   string
	OutputRoot    string
}

type Content struct {
	Config struct {
		Template string
		Items    []string
	}
	Items []*Content
	Data  map[string]any
}

type Processor interface {
	LoadTemplates() (*template.Template, error)
	LoadContent(map[string]*Content) error
	ClearExistingBuild() error
	ProcessContent(*template.Template, *Content, string) error
	CopyStatic() error
}
