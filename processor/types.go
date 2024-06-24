package processor

import "text/template"

type Config struct {
	MappingFile   string
	StaticRoot    string
	TemplatesRoot string
	ContentRoot   string
	OutputRoot    string
}

type RawContentFile struct {
	Item map[string]map[string]any // itemname->item
}

type Item struct {
	Name            string
	RelativeDirPath string
	Data            map[string]any
}

type RawMapping struct {
	OutputBase string
	Template   string
	Selector   string
}

type MappingForTemplate struct {
	MappingPath string
	OutputBase  string
	Template    string
	Selector    string
}

type MappingFile struct {
	Mapping []RawMapping
}

type Processor interface {
	LoadTemplates() (*template.Template, bool)
	LoadContentItems() ([]Item, bool)
	LoadMappings() ([]MappingForTemplate, bool)
	ClearExistingOutput() bool
	ProcessContent(*template.Template, []MappingForTemplate, []Item) bool
	CopyStatic() bool
}
