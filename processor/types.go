package processor

import "io"

type Config struct {
	ContentRoot     string `yaml:"ContentRoot"`
	SiteContentFile string `yaml:"SiteContentFile"`
	MappingFile     string `yaml:"MappingFile"`
	StaticRoot      string `yaml:"StaticRoot"`
	TemplatesRoot   string `yaml:"TemplatesRoot"`
	TemplatesType   string `yaml:"TemplatesType"`
	OutputRoot      string `yaml:"OutputRoot"`
}

type Content map[string]any

type RawMapping struct {
	SingleOutput   string `yaml:"SingleOutput"`
	PerMatchOutput string `yaml:"PerMatchOutput"`
	Template       string `yaml:"Template"`
	Selector       string `yaml:"Selector"`
}

type MappingForTemplate struct {
	SingleOutput   string
	PerMatchOutput string
	Template       string
	Selector       string
}

type Processor interface {
	LoadTemplates() bool
	LoadSiteContent() (any, bool)
	LoadMappings() ([]MappingForTemplate, bool)
	ClearExistingOutput() bool
	ProcessContent([]MappingForTemplate, any) bool
	CopyStatic() bool
}

type TemplateMgr interface {
	ParseOne(tmplName string, tmplBody []byte) error
	Execute(tmplName string, tmplData any, output io.Writer) error
}
