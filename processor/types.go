package processor

type Config struct {
	TemplatesRoot string
	ContentRoot   string
	OutputRoot    string
}

type Content struct {
	Config struct {
		Template string
	}
	Subdatas struct {
		Patterns []string
		Matches  []*Content
	}
	Data map[string]any
}
