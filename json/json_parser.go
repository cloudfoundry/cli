package json

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/cli/cf/errors"
)

func ParseJSON(path string) ([]map[string]string, error) {
	if path == "" {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	stringMaps := []map[string]string{}
	err = json.Unmarshal(bytes, &stringMaps)
	if err != nil {
		return nil, errors.NewWithFmt("Incorrect json format: %s", err.Error())
	}

	return stringMaps, nil
}
