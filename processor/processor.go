package processor

import (
	"bytes"
	"os"
	"path/filepath"
)

type processor struct {
	siteRoot        string
	config          Config
	contentLoader   FileLoader
	templatesLoader FileLoader
	mappingLoader   FileLoader
	staticLoader    FileLoader
	templateMgr     TemplateMgr
}

func Load(
	readFileFn func(string) ([]byte, error),
	configPath string,
	templateMgrFactories map[string]func(string) TemplateMgr,
) (Processor, bool) {

	Printfln("\nLOADING CONFIG FILE...")

	if configPath == "" {
		return nil, Errorfln("--config must be defined")
	}

	configLoader := MakeFileLoader(".", ".", readFileFn)

	var config Config
	err := configLoader.LoadFile(configPath, &config)
	if err != nil {
		return nil, Errorfln("error decoding config file: %s", err.Error())
	}

	// Clean the config
	config.StaticRoot = filepath.Clean(config.StaticRoot)
	config.OutputRoot = filepath.Clean(config.OutputRoot) + "/"

	if config.MappingFile == "" {
		Errorfln("MappingFile must not be empty.")
		return nil, true
	}

	siteRoot := filepath.Dir(configPath) + "/"

	if config.TemplatesType == "" {
		Errorfln("TemplatesType must not be empty.")
		return nil, true
	}
	templateMgrFactory, hasType := templateMgrFactories[config.TemplatesType]
	if !hasType {
		Errorfln("Unrecognized TemplatesType %q.", config.TemplatesType)
		return nil, true
	}

	// TODO(treaster): Consider if data URLs should pull assets relative to
	// siteRoot, or contentRoot? SiteRoot for now I guess.
	templateMgr := templateMgrFactory(siteRoot)

	contentLoader := MakeFileLoader(
		siteRoot,
		config.ContentRoot,
		readFileFn,
	)

	templatesLoader := MakeFileLoader(
		siteRoot,
		config.TemplatesRoot,
		readFileFn,
	)

	staticLoader := MakeFileLoader(
		siteRoot,
		config.StaticRoot,
		readFileFn,
	)

	return &processor{
		siteRoot,
		config,
		contentLoader,
		templatesLoader,
		contentLoader, // contentLoader also works as mappingLoader
		staticLoader,
		templateMgr,
	}, false
}

func (p *processor) LoadTemplates() bool {
	Printfln("\nLOADING TEMPLATES...")

	templateNames := p.templatesLoader.FindFiles()
	if len(templateNames) == 0 {
		return Errorfln("no templates found in templates root %q", p.templatesLoader.BaseDir())
	}

	hasError := false
	for _, templateName := range templateNames {
		tmplContents, err := p.templatesLoader.LoadFileAsBytes(templateName)
		if err != nil {
			newError := Errorfln("error reading template %q: %s", templateName, err.Error())
			hasError = hasError || newError
			continue
		}

		err = p.templateMgr.ParseOne(templateName, tmplContents)
		if err != nil {
			newError := Errorfln("error parsing template %q: %s", templateName, err.Error())
			hasError = hasError || newError
		}
	}

	Printfln("Loaded templates:")
	for _, tmplName := range templateNames {
		Printfln("  %s", tmplName)
	}
	Printfln("")

	return hasError
}

func (p *processor) LoadSiteContent() (any, bool) {
	Printfln("\nLOADING SITE CONTENT...")

	hasError := false
	siteContent, errors := EvalContentFile(p.contentLoader, p.config.SiteContentFile)
	if len(errors) > 0 {
		return nil, false
	}

	return siteContent, hasError
}

func (p *processor) LoadMappings() ([]MappingForTemplate, bool) {
	Printfln("\nLOADING MAPPING FILES...")

	hasError := false
	mappingPaths := p.mappingLoader.FindFilesWithName(p.config.MappingFile)

	var allMappings []MappingForTemplate
	for _, mappingPath := range mappingPaths {
		Printfln("  mapping path %s", mappingPath)
		var rawMappings []RawMapping
		err := p.mappingLoader.LoadFile(mappingPath, &rawMappings)
		if err != nil {
			hasError = Errorfln("error loading mapping file %s: %s", mappingPath, err.Error())
			continue
		}

		Printfln("found mappings %+v", rawMappings)

		for i, rawMapping := range rawMappings {
			Printfln("    found mapping for types %+v onto template %q", rawMapping.Selector, rawMapping.Template)

			AssertNonEmpty(rawMapping.Template)

			if (rawMapping.SingleOutput == "" && rawMapping.PerMatchOutput == "") ||
				(rawMapping.SingleOutput != "" && rawMapping.PerMatchOutput != "") {
				hasError = Errorfln("exactly one of single_output or per_match_output must be set on mapping %s %i", mappingPath, i)
				continue
			}

			forTemplate := MappingForTemplate{
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

func (p *processor) ProcessContent(allMappings []MappingForTemplate, siteContent any) bool {
	Printfln("\nEXECUTING CONTENT + TEMPLATES...")

	hasError := false
	for _, mapping := range allMappings {
		Printfln("    processOneMapping")
		newError := p.processOneMapping(mapping, siteContent)
		hasError = hasError || newError
	}
	return hasError
}

func (p *processor) processOneMapping(mapping MappingForTemplate, siteContent any) bool {
	templateName := mapping.Template
	if templateName == "" {
		Errorfln("mapping file must contain key 'config.template', which defines which template file should be used.")
		return true
	}

	itemMatches := EvalContentExpr(mapping.Selector, siteContent)
	Printfln("SELECTOR %q found %d matches", mapping.Selector, len(itemMatches))

	hasError := false

	if mapping.SingleOutput != "" {
		newError := p.executeOneTemplate(templateName, itemMatches, mapping.SingleOutput)
		hasError = hasError || newError
	}
	if mapping.PerMatchOutput != "" {
		for _, item := range itemMatches {
			itemName := EvalOutputBase(mapping.PerMatchOutput, item)
			newError := p.executeOneTemplate(templateName, item, itemName)
			hasError = hasError || newError
		}
	}

	return hasError
}

func (p *processor) executeOneTemplate(tmplName string, tmplData any, outputRelPath string) bool {
	Printfln("Execute template %s", tmplName)

	var output bytes.Buffer
	err := p.templateMgr.Execute(tmplName, tmplData, &output)
	if err != nil {
		return Errorfln("error executing template: %s", err.Error())
	}

	outputPath := filepath.Join(p.siteRoot, p.config.OutputRoot, outputRelPath)
	outputDir := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return Errorfln("error creating output directory: %s", err.Error())
	}

	Printfln("    Writing file %s", outputPath)
	err = os.WriteFile(outputPath, output.Bytes(), 0644)
	if err != nil {
		return Errorfln("error writing output file: %s", err.Error())
	}

	return false
}

func (p *processor) CopyStatic() bool {
	Printfln("\nCOPYING STATIC FILES...")

	hasError := false

	staticFiles := p.staticLoader.FindFiles()
	Printfln("copying %d static files", len(staticFiles))
	for _, staticFile := range staticFiles {
		outPath := filepath.Join(p.siteRoot, p.config.OutputRoot, p.config.StaticRoot, staticFile)
		outDir := filepath.Dir(outPath)
		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			Printfln("error making dir for static files: %s", outDir, err.Error())
			hasError = true
			continue
		}

		Printfln("    copy %s to %s", staticFile, outPath)
		err = p.staticLoader.Copy(staticFile, outPath)
		if err != nil {
			Printfln("error copying file %s to %s: %s", staticFile, outPath, err.Error())
			hasError = true
			continue
		}
	}

	return hasError
}
