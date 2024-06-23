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
	Item map[string]map[string]map[string]any // type->itemname->item
}

type Item struct {
	Name            string
	Type            string
	RelativeDirPath string
	Data            map[string]any
}

type RawMapping struct {
	MappingType string
	OutputBase  string
	Template    string
	ItemTypes   []string
	SortKey     string
}

type MappingForTemplate struct {
	MappingPath string
	MappingType string
	OutputBase  string
	Template    string
	ItemTypes   []string
	SortKey     string
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
