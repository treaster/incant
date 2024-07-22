package processor

import "text/template"

type Config struct {
	MappingFile   string `yaml:"MappingFile"`
	StaticRoot    string `yaml:"StaticRoot"`
	TemplatesRoot string `yaml:"TemplatesRoot"`
	SiteDataFile  string `yaml:"SiteDataFile"`
	OutputRoot    string `yaml:"OutputRoot"`
}

type Content map[string]any

type RawMapping struct {
	SingleOutput   string
	PerMatchOutput string
	Template       string
	Selector       string
}

type MappingForTemplate struct {
	MappingPath    string
	SingleOutput   string
	PerMatchOutput string
	Template       string
	Selector       string
}

type MappingFile struct {
	Mapping []RawMapping
}

type Processor interface {
	LoadTemplates() (*template.Template, bool)
	LoadSiteContent() (any, bool)
	LoadMappings() ([]MappingForTemplate, bool)
	ClearExistingOutput() bool
	ProcessContent(*template.Template, []MappingForTemplate, any) bool
	CopyStatic() bool
}
