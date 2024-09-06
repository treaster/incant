package processor

import "text/template"

type Config struct {
	ContentRoot   string `yaml:"ContentRoot"`
	SiteDataFile  string `yaml:"SiteDataFile"`
	MappingFile   string `yaml:"MappingFile"`
	StaticRoot    string `yaml:"StaticRoot"`
	TemplatesRoot string `yaml:"TemplatesRoot"`
	OutputRoot    string `yaml:"OutputRoot"`
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
	LoadTemplates() (*template.Template, bool)
	LoadSiteContent() (any, bool)
	LoadMappings() ([]MappingForTemplate, bool)
	ClearExistingOutput() bool
	ProcessContent(*template.Template, []MappingForTemplate, any) bool
	CopyStatic() bool
}
