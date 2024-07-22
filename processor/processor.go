package processor

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/treaster/ssg/processor/content_file"
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
	config.SiteDataFile = filepath.Clean(config.SiteDataFile) + "/"
	config.OutputRoot = filepath.Clean(config.OutputRoot) + "/"

	if config.MappingFile == "" || filepath.Base(config.MappingFile) != config.MappingFile {
		Errorfln("MappingFile must be a nonempty, bare filename. No directory or path separators should be included. Got %q.", config.MappingFile)
		return nil, true
	}

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

		tmplName := SafeCutPrefix(oneTmpl, filepath.Join(p.siteRoot, p.config.TemplatesRoot)+"/")
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

func (p *processor) LoadSiteContent() (map[string]any, bool) {
	Printfln("\nLOADING SITE CONTENT...")

	hasError := false
	contentRootToml := filepath.Join(p.siteRoot, p.config.SiteDataFile)
	siteData, errors := content_file.EvalContentFile(os.ReadFile, contentRootToml)
	if len(errors) > 0 {
		return nil, false
	}

	return siteData, hasError
}

func (p *processor) LoadMappings() ([]MappingForTemplate, bool) {
	Printfln("\nLOADING MAPPING FILES...")

	hasError := false
	mappingPaths := FindFilesWithName(filepath.Join(p.siteRoot, p.config.SiteDataFile), p.config.MappingFile)

	var allMappings []MappingForTemplate
	for _, mappingPath := range mappingPaths {
		Printfln("  mapping path %s", mappingPath)
		var rawMappings MappingFile
		_, err := toml.DecodeFile(mappingPath, &rawMappings)
		if err != nil {
			hasError = Errorfln("error loading mapping file: %s", err.Error())
			continue
		}

		cleanPath := SafeCutPrefix(mappingPath, filepath.Join(p.siteRoot, p.config.SiteDataFile))
		cleanPath, _ = TrimExt(cleanPath)

		for i, rawMapping := range rawMappings.Mapping {
			Printfln("    found mapping for types %+v onto template %q", rawMapping.Selector, rawMapping.Template)

			AssertNonEmpty(rawMapping.Template)

			if (rawMapping.SingleOutput == "" && rawMapping.PerMatchOutput == "") ||
				(rawMapping.SingleOutput != "" && rawMapping.PerMatchOutput != "") {
				hasError = Errorfln("exactly one of single_output or per_match_output must be set on mapping %s %i", mappingPath, i)
				continue
			}

			forTemplate := MappingForTemplate{
				cleanPath,
				rawMapping.SingleOutput,
				rawMapping.PerMatchOutput,
				rawMapping.Template,
				rawMapping.Selector,
			}
			allMappings = append(allMappings, forTemplate)
		}
	}

	return allMappings, hasError
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

func (p *processor) ProcessContent(tmpl *template.Template, allMappings []MappingForTemplate, siteContent map[string]any) bool {
	Printfln("\nEXECUTING CONTENT + TEMPLATES...")

	hasError := false
	for _, mapping := range allMappings {
		newError := p.processOneMapping(tmpl, mapping, siteContent)
		hasError = hasError || newError
	}
	return hasError
}

func (p *processor) processOneMapping(tmpl *template.Template, mapping MappingForTemplate, siteContent map[string]any) bool {
	templateName := mapping.Template
	if templateName == "" {
		Errorfln("mapping file must contain key 'config.template', which defines which template file should be used.")
		return true
	}

	oneTmpl := tmpl.Lookup(templateName)
	if oneTmpl == nil {
		panic(fmt.Sprintf("error: template %q not found", templateName))
	}

	itemMatches := FilterBySelector(mapping.Selector, siteContent)

	hasError := false

	Printfln("selected %d items", len(itemMatches))
	if mapping.SingleOutput != "" {
		newError := p.executeOneTemplate(oneTmpl, itemMatches, mapping.SingleOutput)
		hasError = hasError || newError
	}
	if mapping.PerMatchOutput != "" {
		for _, item := range itemMatches {
			itemName := EvalOutputBase(mapping.PerMatchOutput, item)
			newError := p.executeOneTemplate(oneTmpl, item, itemName)
			hasError = hasError || newError
		}
	}

	return hasError
}

func (p *processor) executeOneTemplate(tmpl *template.Template, tmplData any, outputBase string) bool {
	Printfln("Execute template %s", tmpl.Name())

	var output bytes.Buffer
	err := tmpl.Execute(&output, tmplData)
	if err != nil {
		return Errorfln("error executing template: %s", err.Error())
	}

	tmplExt := filepath.Ext(tmpl.Name())

	outputPath := filepath.Join(p.siteRoot, p.config.OutputRoot, outputBase+tmplExt)
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
		partialPath := SafeCutPrefix(staticFile, prefix)

		outPath := filepath.Join(p.siteRoot, p.config.OutputRoot, p.config.StaticRoot, partialPath)
		outDir := filepath.Dir(outPath)
		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			Printfln("error making dir for static files: %s", outDir, err.Error())
			hasError = true
			continue
		}

		err = Copy(staticFile, outPath)
		if err != nil {
			Printfln("error copying file %s to %s: %s", staticFile, outPath, err.Error())
			hasError = true
			continue
		}
	}

	return hasError
}
