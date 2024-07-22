package main

import (
	"flag"

	"github.com/treaster/ssg/processor"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "YAML file defining static site params")

	flag.Parse()

	proc, hasErrors := processor.Load(configPath)
	if hasErrors {
		processor.Printfln("ERROR loading config")
		return
	}

	tmpl, hasError := proc.LoadTemplates()
	if hasError {
		processor.Printfln("ERROR loading template files")
		return
	}

	siteContent, hasErrors := proc.LoadSiteContent()
	if hasErrors {
		processor.Printfln("ERROR loading site content")
		return
	}

	allMappings, hasErrors := proc.LoadMappings()
	if hasErrors {
		processor.Printfln("ERROR loading mapping files")
		return
	}

	hasErrors = proc.ClearExistingOutput()
	if hasErrors {
		processor.Printfln("ERROR clearing existing output")
		return
	}

	hasErrors = proc.ProcessContent(tmpl, allMappings, siteContent)
	if hasErrors {
		processor.Printfln("ERROR processing mapping + site content")
		return
	}

	hasErrors = proc.CopyStatic()
	if hasErrors {
		processor.Printfln("ERROR copying static files")
		return
	}
}
