package main

import (
	gojson "encoding/json"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/cli/json"
)

func main() {
	pathToJSONFile := os.Args[1]
	rules, err := json.ParseJSON(pathToJSONFile)
	if err != nil {
		panic(err.Error())
	}

	jsonOutput, jsonErr := gojson.MarshalIndent(rules, "", "   ")
	if jsonErr != nil {
		panic(jsonErr.Error())
	}
	ioutil.WriteFile(pathToJSONFile, jsonOutput, 0644)
}
