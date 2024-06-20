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
	LoadTemplates() (*template.Template, bool)
	LoadContent(map[string]*Content) bool
	ClearExistingOutput() bool
	ProcessContent(*template.Template, map[string]*Content) bool
	CopyStatic() bool
}
