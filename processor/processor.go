package processor

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

type Processor interface {
	LoadTemplates() (*template.Template, error)
	LoadContent(map[string]*Content) error
	ProcessContent(*template.Template, *Content, string) error
}

type processor struct {
	configPath string
	siteRoot   string
	config     Config
}

func Load(configPath string) (Processor, error) {
	var config Config

	if configPath == "" {
		return nil, fmt.Errorf("--config must be defined")
	}

	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return nil, err
	}

	siteRoot := filepath.Dir(configPath) + "/"
	return &processor{
		configPath,
		siteRoot,
		config,
	}, nil
}

func (p *processor) LoadTemplates() (*template.Template, error) {
	tmpl := template.
		New("ssg").
		Funcs(template.FuncMap{
			"RenderMarkdown": RenderMarkdown,
		}).
		Option("missingkey=error")

	templates := FindFiles(filepath.Join(p.siteRoot, p.config.TemplatesRoot))
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found in templates root %q", p.config.TemplatesRoot)
	}

	var templateNames []string
	hasError := false
	for _, oneTmpl := range templates {
		tmplContents, err := os.ReadFile(oneTmpl)
		if err != nil {
			fmt.Errorf("error reading template %q: %s", oneTmpl, err.Error())
			hasError = true
		}

		tmplName, isOk := strings.CutPrefix(oneTmpl, filepath.Join(p.siteRoot, p.config.TemplatesRoot)+"/")
		if !isOk {
			panic(fmt.Sprintf("Error removing prefix %q on %q", p.siteRoot, oneTmpl))
		}
		_, err = tmpl.New(tmplName).Parse(string(tmplContents))
		if err != nil {
			fmt.Errorf("error parsing template %q: %s", oneTmpl, err.Error())
			hasError = true
		}

		templateNames = append(templateNames, tmplName)
	}

	if hasError {
		return nil, fmt.Errorf("encountered errors while parsing templates")
	}

	Printfln("Loaded templates:")
	for _, tmplName := range templateNames {
		Printfln("  %s", tmplName)
	}
	Printfln("")

	return tmpl, nil
}

func (p *processor) LoadContent(output map[string]*Content) error {
	hasError := false
	contentFiles := FindFiles(filepath.Join(p.siteRoot, p.config.ContentRoot))
	for _, contentPath := range contentFiles {
		Printfln("Processing content file: %s...", contentPath)
		var content Content

		metadata, err := toml.DecodeFile(contentPath, &content)
		if err != nil {
			Printfln("  error decoding TOML: %s", err.Error())
			continue
		}

		var toplevelKeys []string
		for _, key := range metadata.Keys() {
			if strings.IndexRune(key.String(), '.') >= 0 {
				continue
			}
			toplevelKeys = append(toplevelKeys, key.String())
		}

		if slices.Index(toplevelKeys, "config") < 0 {
			Printfln("  missing required key 'config'")
			hasError = true
			continue
		}
		if slices.Index(toplevelKeys, "data") < 0 {
			Printfln("  missing required key 'data'")
			hasError = true
			continue
		}
		if len(toplevelKeys) > 2 {
			Printfln("  expected exactly two top-level keys, 'config' and 'data'. found %+v", toplevelKeys)
			hasError = true
			continue
		}

		contentPath, hasPrefix := strings.CutPrefix(contentPath, p.siteRoot)
		if !hasPrefix {
			panic("Whaaa?")
		}
		output[contentPath] = &content
	}

	for thisContentPath, thisContent := range output {
		for _, subdataPattern := range thisContent.Config.Items {
			for candidatePath, candidateContent := range output {
				if candidatePath == thisContentPath {
					continue
				}

				Printfln("IsMatch(%q, %q)", subdataPattern, candidatePath)
				isMatch, err := filepath.Match(subdataPattern, candidatePath)
				if err != nil {
					Printfln("error loading %s: %s", thisContentPath, err.Error())
					hasError = true
					continue
				}
				if isMatch {
					thisContent.Items = append(thisContent.Items, candidateContent)
				}
			}
		}
	}

	if hasError {
		return fmt.Errorf("encountered errors while processing content files.")
	}

	for k, v := range output {
		Printfln("%q: %+v", k, *v)
	}

	return nil
}

func (p *processor) ProcessContent(tmpl *template.Template, content *Content, contentPath string) error {
	templateName := content.Config.Template
	if templateName == "" {
		return fmt.Errorf("content file must contain key 'config.template', which defines which template file should be used.")
	}

	var output bytes.Buffer
	oneTmpl := tmpl.Lookup(templateName)
	err := oneTmpl.Execute(&output, content)
	if err != nil {
		return err
	}

	tmplExt := filepath.Ext(templateName)

	relative, isOk := strings.CutPrefix(filepath.Clean(contentPath), filepath.Clean(p.config.ContentRoot))
	if !isOk {
		panic("Whaa?")
	}

	contentExt := filepath.Ext(relative)
	relativeNoExt, isOk := strings.CutSuffix(relative, contentExt)
	if !isOk {
		panic("Whaa?")
	}

	relative = relativeNoExt + tmplExt

	outputPath := filepath.Join(p.siteRoot, p.config.OutputRoot, relative)
	outputDir := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, output.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %s", err.Error())
	}

	return nil
}
