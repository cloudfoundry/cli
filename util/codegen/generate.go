// +build ignore

package main

import (
	"fmt"
	"os"
	"text/template"

	"code.cloudfoundry.org/cli/util/codegen"
)

func main() {
	entity := os.Args[1]
	templatePath := os.Args[2]
	outputPath := os.Args[3]

	template, err := template.ParseFiles(templatePath)
	if err != nil {
		panic(err)
	}

	outputFile, err := os.Create(outputPath)
	defer outputFile.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("generating %s from %s\n", outputPath, templatePath)

	fmt.Fprintf(outputFile, "// generated from %s\n\n", templatePath)

	templateInput := codegen.NewTemplateInput(entity)
	err = template.Execute(outputFile, templateInput)
	if err != nil {
		panic(err)
	}
}
