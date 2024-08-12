package main

import (
	"flag"
	"os"

	"github.com/treaster/incant/processor"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "YAML file defining static site params")

	flag.Parse()

	proc, hasErrors := processor.Load(os.ReadFile, configPath)
	if hasErrors {
		processor.Printfln("ERROR loading config")
		os.Exit(1)
	}

	tmpl, hasError := proc.LoadTemplates()
	if hasError {
		processor.Printfln("ERROR loading template files")
		os.Exit(1)
	}

	siteContent, hasErrors := proc.LoadSiteContent()
	if hasErrors {
		processor.Printfln("ERROR loading site content")
		os.Exit(1)
	}
	if siteContent == nil {
		processor.Printfln("ERROR site content is nil.")
		os.Exit(1)
	}

	allMappings, hasErrors := proc.LoadMappings()
	if hasErrors {
		processor.Printfln("ERROR loading mapping files")
		os.Exit(1)
	}
	if len(allMappings) == 0 {
		processor.Printfln("No mapping files found. Aborting.")
		os.Exit(1)
	}

	hasErrors = proc.ClearExistingOutput()
	if hasErrors {
		processor.Printfln("ERROR clearing existing output")
		os.Exit(1)
	}

	hasErrors = proc.ProcessContent(tmpl, allMappings, siteContent)
	if hasErrors {
		processor.Printfln("ERROR processing mapping + site content")
		os.Exit(1)
	}

	hasErrors = proc.CopyStatic()
	if hasErrors {
		processor.Printfln("ERROR copying static files")
		os.Exit(1)
	}
}
