package processor

import "text/template"

type Config struct {
	MappingFile   string
	StaticRoot    string
	TemplatesRoot string
	SiteDataFile  string
	OutputRoot    string
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
	LoadSiteContent() (map[string]any, bool)
	LoadMappings() ([]MappingForTemplate, bool)
	ClearExistingOutput() bool
	ProcessContent(*template.Template, []MappingForTemplate, map[string]any) bool
	CopyStatic() bool
}
