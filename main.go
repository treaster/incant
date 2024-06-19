package main

import (
	"flag"
	"fmt"

	"github.com/treaster.net/ssg/processor"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "TOML file defining static site params")

	flag.Parse()

	proc, err := processor.Load(configPath)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}

	allContents := map[string]*processor.Content{}

	tmpl, err := proc.LoadTemplates()
	if err != nil {
		fmt.Printf("ERROR: error loading template files: %s\n", err.Error())
		return
	}

	err = proc.LoadContent(allContents)
	if err != nil {
		fmt.Printf("ERROR: error loading content files: %s\n", err.Error())
		return
	}

	hasError := false
	for contentPath, content := range allContents {
		err := proc.ProcessContent(tmpl, content, contentPath)
		if err != nil {
			hasError = true
			fmt.Printf("ERROR: error processing content %q: %s\n", contentPath, err.Error())
		}
	}
	if hasError {
		return
	}
}
