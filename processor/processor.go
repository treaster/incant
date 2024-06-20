package processor

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

type processor struct {
	configPath string
	siteRoot   string
	config     Config
}

func Load(configPath string) (Processor, bool) {
	Printfln("\nLOADING CONFIG FILE...")

	var config Config

	if configPath == "" {
		return nil, Errorfln("--config must be defined")
	}

	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return nil, Errorfln("error decoding config file: %s", err.Error())
	}

	config.StaticRoot = filepath.Clean(config.StaticRoot) + "/"
	config.TemplatesRoot = filepath.Clean(config.TemplatesRoot) + "/"
	config.ContentRoot = filepath.Clean(config.ContentRoot) + "/"
	config.OutputRoot = filepath.Clean(config.OutputRoot) + "/"

	siteRoot := filepath.Dir(configPath) + "/"
	return &processor{
		configPath,
		siteRoot,
		config,
	}, false
}

func (p *processor) LoadTemplates() (*template.Template, bool) {
	Printfln("\nLOADING TEMPLATES...")

	tmpl := template.
		New("ssg").
		Funcs(template.FuncMap{
			"RenderMarkdown": RenderMarkdown,
			"DataUrl": func(assetType string, assetPath string) string {
				fullPath := filepath.Join(p.siteRoot, assetPath)
				return DataUrl(assetType, fullPath)
			},
		}).
		Option("missingkey=error")

	templates := FindFiles(filepath.Join(p.siteRoot, p.config.TemplatesRoot))
	if len(templates) == 0 {
		return nil, Errorfln("no templates found in templates root %q", p.config.TemplatesRoot)
	}

	var templateNames []string
	hasError := false
	for _, oneTmpl := range templates {
		tmplContents, err := os.ReadFile(oneTmpl)
		if err != nil {
			newError := Errorfln("error reading template %q: %s", oneTmpl, err.Error())
			hasError = hasError || newError
			continue
		}

		tmplName, isOk := strings.CutPrefix(oneTmpl, filepath.Join(p.siteRoot, p.config.TemplatesRoot)+"/")
		if !isOk {
			panic(fmt.Sprintf("Error removing prefix %q on %q", p.siteRoot, oneTmpl))
		}
		_, err = tmpl.New(tmplName).Parse(string(tmplContents))
		if err != nil {
			newError := Errorfln("error parsing template %q: %s", oneTmpl, err.Error())
			hasError = hasError || newError
			continue
		}

		templateNames = append(templateNames, tmplName)
	}

	Printfln("Loaded templates:")
	for _, tmplName := range templateNames {
		Printfln("  %s", tmplName)
	}
	Printfln("")

	return tmpl, hasError
}

func (p *processor) LoadContent(output map[string]*Content) bool {
	Printfln("\nLOADING CONTENT FILES...")

	hasError := false
	contentFiles := FindFiles(filepath.Join(p.siteRoot, p.config.ContentRoot))

	for _, contentPath := range contentFiles {
		Printfln("Processing content file: %s...", contentPath)
		var rawContent Content

		metadata, err := toml.DecodeFile(contentPath, &rawContent)
		if err != nil {
			newError := Errorfln("  error decoding TOML: %s", err.Error())
			hasError = hasError || newError
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
			newError := Errorfln("  missing required key 'config'")
			hasError = hasError || newError
			continue
		}
		if slices.Index(toplevelKeys, "data") < 0 {
			newError := Errorfln("  missing required key 'data'")
			hasError = hasError || newError
			continue
		}
		if len(toplevelKeys) > 2 {
			newError := Errorfln("  expected exactly two top-level keys, 'config' and 'data'. found %+v", toplevelKeys)
			hasError = hasError || newError
			continue
		}

		contentPath, hasPrefix := strings.CutPrefix(contentPath, p.siteRoot)
		if !hasPrefix {
			panic("Whaaa?")
		}

		tmplExt := filepath.Ext(rawContent.Config.Template)
		contentExt := filepath.Ext(contentPath)
		contentPathNoExt, isOk := strings.CutSuffix(contentPath, contentExt)
		if !isOk {
			panic("Whaa?")
		}
		outputPath := contentPathNoExt + tmplExt
		outputPath, isOk = strings.CutPrefix(outputPath, p.config.ContentRoot)
		if !isOk {
			panic(fmt.Sprintf("Whaa? %s vs %s", outputPath, p.config.ContentRoot))
		}

		rawContent.contentPath = contentPath
		rawContent.forTemplate = &ForTemplate{
			Path: outputPath,
			Data: rawContent.Data,
		}

		output[contentPath] = &rawContent
	}

	if hasError {
		return hasError
	}

	for _, content := range output {
		for _, subdataPattern := range content.Config.Items {
			for candidatePath, candidateContent := range output {
				if candidatePath == content.contentPath {
					continue
				}

				Printfln("IsMatch(%q, %q)", subdataPattern, candidatePath)
				isMatch, err := filepath.Match(subdataPattern, candidatePath)
				if err != nil {
					newError := Errorfln("error loading %s: %s", content.forTemplate.Path, err.Error())
					hasError = hasError || newError
					continue
				}
				if isMatch {
					content.forTemplate.Items = append(content.forTemplate.Items, candidateContent.forTemplate)
				}
			}
		}
	}

	for k, v := range output {
		Printfln("  %q: %+v", k, *v)
	}

	return hasError
}

func (p *processor) ClearExistingOutput() bool {
	Printfln("\nCLEARING EXISTING OUTPUT...")

	outputPath := filepath.Join(p.siteRoot, p.config.OutputRoot)
	err := os.RemoveAll(outputPath)
	if err != nil {
		return Errorfln("error deleting existing output: %s", err.Error())
	}
	return false
}

func (p *processor) ProcessContent(tmpl *template.Template, allContents map[string]*Content) bool {
	Printfln("\nEXECUTING CONTENT + TEMPLATES...")

	hasError := false
	for contentPath, content := range allContents {
		newError := p.processOneContent(tmpl, content, contentPath)
		hasError = hasError || newError
	}

	return hasError
}

func (p *processor) processOneContent(tmpl *template.Template, content *Content, contentPath string) bool {
	templateName := content.Config.Template
	if templateName == "" {
		return Errorfln("content file must contain key 'config.template', which defines which template file should be used.")
	}

	var output bytes.Buffer
	oneTmpl := tmpl.Lookup(templateName)

	Printfln("CONTENT %+v", content.forTemplate)
	err := oneTmpl.Execute(&output, content.forTemplate)
	if err != nil {
		return Errorfln("error executing template: %s", err.Error())
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
		return Errorfln("error creating output directory: %s", err.Error())
	}

	err = os.WriteFile(outputPath, output.Bytes(), 0644)
	if err != nil {
		return Errorfln("error writing output file: %s", err.Error())
	}

	return false
}

func (p *processor) CopyStatic() bool {
	Printfln("\nCOPYING STATIC FILES...")

	hasError := false

	prefix := filepath.Join(p.siteRoot, p.config.StaticRoot)
	staticFiles := FindFiles(prefix)
	for _, staticFile := range staticFiles {
		partialPath, hasPrefix := strings.CutPrefix(staticFile, prefix)
		if !hasPrefix {
			panic("Whaaa?")
		}

		outPath := filepath.Join(p.siteRoot, p.config.OutputRoot, p.config.StaticRoot, partialPath)
		outDir := filepath.Dir(outPath)
		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			Printfln("error making dir for static files: %s", outDir, err.Error())
			hasError = true
		}

		err = Copy(staticFile, outPath)
		if err != nil {
			Printfln("error copying file %s to %s: %s", staticFile, outPath, err.Error())
			hasError = true
		}
	}

	return hasError
}
