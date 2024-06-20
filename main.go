package main

import (
	"flag"

	"github.com/treaster.net/ssg/processor"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "TOML file defining static site params")

	flag.Parse()

	proc, hasErrors := processor.Load(configPath)
	if hasErrors {
		processor.Printfln("ERROR loading config")
		return
	}

	allContents := map[string]*processor.Content{}

	tmpl, hasError := proc.LoadTemplates()
	if hasError {
		processor.Printfln("ERROR loading template files")
		return
	}

	hasErrors = proc.LoadContent(allContents)
	if hasErrors {
		processor.Printfln("ERROR loading content files")
		return
	}

	hasErrors = proc.ClearExistingOutput()
	if hasErrors {
		processor.Printfln("ERROR clearing existing output")
		return
	}

	hasErrors = proc.ProcessContent(tmpl, allContents)
	if hasErrors {
		processor.Printfln("ERROR processing content files")
		return
	}

	hasErrors = proc.CopyStatic()
	if hasErrors {
		processor.Printfln("ERROR loading content files")
		return
	}
}
