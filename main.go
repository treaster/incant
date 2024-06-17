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

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "TOML file defining static site params")

	flag.Parse()

	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}

	tmpl, err := LoadTemplates(config)

	hasError := false
	contents := FindFiles(config.ContentRoot)
	for _, file := range contents {
		err := ProcessContent(config, tmpl, file)
		if err != nil {
			hasError = true
			fmt.Printf("ERROR: error processing content %q: %s\n", file, err.Error())
		}
	}
	if hasError {
		return
	}
}

type Config struct {
	TemplatesRoot string
	ContentRoot   string
	OutputRoot    string
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

func LoadTemplates(config Config) (*template.Template, error) {
	tmpl := template.
		New("ssg").
		Funcs(template.FuncMap{
			"RenderMarkdown": RenderMarkdown,
		}).
		Option("missingkey=error")

	templates := FindFiles(config.TemplatesRoot)
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found in templates root %q\n", config.TemplatesRoot)
	}

	fmt.Println("parse templates", templates)
	hasError := false
	for _, oneTmpl := range templates {
		tmplContents, err := os.ReadFile(oneTmpl)
		if err == nil {
			_, err = tmpl.New(oneTmpl).Parse(string(tmplContents))
		}

		if err != nil {
			fmt.Errorf("error parsing template %q: %s", oneTmpl, err.Error())
			hasError = true
		}
	}

	if hasError {
		return nil, fmt.Errorf("encountered errors while parsing templates")
	}

	fmt.Println(tmpl.DefinedTemplates())
	return tmpl, nil
}

func ProcessContent(config Config, tmpl *template.Template, contentPath string) error {
	var content map[string]any
	_, err := toml.DecodeFile(contentPath, &content)
	if err != nil {
		return err
	}

	templateAny, hasTemplate := content["template"]
	if !hasTemplate {
		return fmt.Errorf("content file must contain key 'template', which defines which template file should be used.")
	}

	templateName, isString := templateAny.(string)
	if !isString {
		return fmt.Errorf("template key must be a string.")
	}
	var output bytes.Buffer
	oneTmpl := tmpl.Lookup(templateName)
	err = oneTmpl.Execute(&output, content)
	if err != nil {
		return err
	}

	relative, isOk := strings.CutPrefix(filepath.Clean(contentPath), filepath.Clean(config.ContentRoot))
	if !isOk {
		panic("Whaa?")
	}

	outputPath := filepath.Join(config.OutputRoot, relative)
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
