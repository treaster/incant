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

	tmpl, err := LoadTemplates(filepath.Dir(configPath)+"/", config)

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
	if err != nil {
		return config, err
	}

	configDir := filepath.Dir(configPath)
	RewritePath(configDir, &config.TemplatesRoot)
	RewritePath(configDir, &config.ContentRoot)
	RewritePath(configDir, &config.OutputRoot)

	return config, err
}

func RewritePath(configDir string, relativePath *string) {
	if filepath.IsAbs(*relativePath) {
		panic(fmt.Sprintf("Paths in config file must be relative to the config file's directory. Absolute paths are not supported."))
	}
	*relativePath = filepath.Join(configDir, *relativePath)
	fmt.Println(*relativePath)
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

func LoadTemplates(configPath string, config Config) (*template.Template, error) {
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
			tmplName, isOk := strings.CutPrefix(oneTmpl, configPath)
			if !isOk {
				panic(fmt.Sprintf("Error removing prefix %q on %q", configPath, oneTmpl))
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

func ProcessContent(config Config, tmpl *template.Template, contentPath string) error {
	var content struct {
		Config struct {
			Template string
		}
		Data map[string]any
	}

	_, err := toml.DecodeFile(contentPath, &content)
	if err != nil {
		return err
	}

	templateName := content.Config.Template
	if templateName == "" {
		return fmt.Errorf("content file must contain key 'config.template', which defines which template file should be used.")
	}

	var output bytes.Buffer
	oneTmpl := tmpl.Lookup(templateName)
	err = oneTmpl.Execute(&output, content)
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
