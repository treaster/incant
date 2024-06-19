package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/yuin/goldmark"
)

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

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "TOML file defining static site params")

	flag.Parse()

	siteRoot := filepath.Dir(configPath) + "/"
	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}

	allContents := map[string]*Content{}

	tmpl, err := LoadTemplates(siteRoot, config)
	if err != nil {
		fmt.Printf("ERROR: error loading template files: %s\n", err.Error())
		return
	}

	err = LoadContent(siteRoot, config, allContents)
	if err != nil {
		fmt.Printf("ERROR: error loading content files: %s\n", err.Error())
		return
	}

	hasError := false
	for contentPath, content := range allContents {
		err := ProcessContent(siteRoot, config, tmpl, content, contentPath)
		if err != nil {
			hasError = true
			fmt.Printf("ERROR: error processing content %q: %s\n", contentPath, err.Error())
		}
	}
	if hasError {
		return
	}
}

func LoadConfig(configPath string) (Config, error) {
	var config Config

	if configPath == "" {
		return config, fmt.Errorf("--config must be defined")
	}

	_, err := toml.DecodeFile(configPath, &config)
	return config, err
}

func FindFiles(fileRoot string) []string {
	var files []string
	filepath.WalkDir(fileRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err.Error())
		}
		baseName := filepath.Base(path)
		if baseName[0] == '.' {
			fmt.Printf("skipping dotfile %s\n", path)
			return nil
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files
}

func LoadTemplates(siteRoot string, config Config) (*template.Template, error) {
	tmpl := template.
		New("ssg").
		Funcs(template.FuncMap{
			"RenderMarkdown": RenderMarkdown,
			//"PageItems":      GetPageItems,
		}).
		Option("missingkey=error")

	templates := FindFiles(filepath.Join(siteRoot, config.TemplatesRoot))
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found in templates root %q\n", config.TemplatesRoot)
	}

	fmt.Println("parse templates", templates)
	hasError := false
	for _, oneTmpl := range templates {
		tmplContents, err := os.ReadFile(oneTmpl)
		if err == nil {
			tmplName, isOk := strings.CutPrefix(oneTmpl, siteRoot)
			if !isOk {
				panic(fmt.Sprintf("Error removing prefix %q on %q", siteRoot, oneTmpl))
			}
			_, err = tmpl.New(tmplName).Parse(string(tmplContents))
		}

		if err != nil {
			fmt.Errorf("error parsing template %q: %s", oneTmpl, err.Error())
			hasError = true
		}
	}

	if hasError {
		return nil, fmt.Errorf("encountered errors while parsing templates")
	}

	fmt.Println("Loaded templates:", tmpl.DefinedTemplates())
	return tmpl, nil
}

func LoadContent(siteRoot string, config Config, output map[string]*Content) error {
	contentFiles := FindFiles(filepath.Join(siteRoot, config.ContentRoot))
	for _, contentPath := range contentFiles {
		var content Content

		_, err := toml.DecodeFile(contentPath, &content)
		if err != nil {
			return err
		}

		contentPath, hasPrefix := strings.CutPrefix(contentPath, siteRoot)
		if !hasPrefix {
			panic("Whaaa?")
		}
		output[contentPath] = &content
	}

	hasError := false
	for thisContentPath, thisContent := range output {
		for _, subdataPattern := range thisContent.Subdatas.Patterns {
			for candidatePath, candidateContent := range output {
				if candidatePath == thisContentPath {
					continue
				}

				fmt.Printf("IsMatch(%q, %q)\n", subdataPattern, candidatePath)
				isMatch, err := filepath.Match(subdataPattern, candidatePath)
				if err != nil {
					fmt.Printf("error loading %s: %s\n", thisContentPath, err.Error())
					hasError = true
					continue
				}
				if isMatch {
					thisContent.Subdatas.Matches = append(thisContent.Subdatas.Matches, candidateContent)
				}
			}
		}
	}

	if hasError {
		return fmt.Errorf("encountered errors while processing content files.")
	}

	for k, v := range output {
		fmt.Printf("%q: %+v\n", k, *v)
	}

	return nil
}

func ProcessContent(siteRoot string, config Config, tmpl *template.Template, content *Content, contentPath string) error {
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

	relative, isOk := strings.CutPrefix(filepath.Clean(contentPath), filepath.Clean(config.ContentRoot))
	if !isOk {
		panic("Whaa?")
	}

	contentExt := filepath.Ext(relative)
	relativeNoExt, isOk := strings.CutSuffix(relative, contentExt)
	if !isOk {
		panic("Whaa?")
	}

	relative = relativeNoExt + tmplExt

	outputPath := filepath.Join(siteRoot, config.OutputRoot, relative)
	outputDir := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, output.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %s\n", err.Error())
	}

	return nil
}

func RenderMarkdown(input string) (string, error) {
	var buf bytes.Buffer
	err := goldmark.Convert([]byte(input), &buf)
	return buf.String(), err
}
