package processor

type Config struct {
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
