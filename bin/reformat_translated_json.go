package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Entry struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: reformat_translated_json <tranlation directory>\n")
		os.Exit(1)
	}
	directory := os.Args[1]
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".all.json") {
			continue
		}
		fullPath := filepath.Join(directory, file.Name())

		fmt.Println("reformatting:", file.Name())

		raw, err := ioutil.ReadFile(fullPath)
		if err != nil {
			panic(err)
		}

		var entries []Entry
		err = json.Unmarshal(raw, &entries)
		if err != nil {
			panic(err)
		}

		rawOut, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(fullPath, rawOut, file.Mode())
		if err != nil {
			panic(err)
		}
	}
}
