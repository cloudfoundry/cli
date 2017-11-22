package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/cli-plugin-repo/sort/yamlsorter"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "must provide a path to yaml file to sort")
		os.Exit(1)
	}
	path := os.Args[1]

	var yamlSorter yamlsorter.YAMLSorter

	unsortedBytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sortedBytes, err := yamlSorter.Sort(unsortedBytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(path, sortedBytes, 0664)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
