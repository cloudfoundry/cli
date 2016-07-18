package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ParseJSONArray(path string) ([]map[string]interface{}, error) {
	if path == "" {
		return nil, nil
	}

	bytes, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	stringMaps := []map[string]interface{}{}
	err = json.Unmarshal(bytes, &stringMaps)
	if err != nil {
		return nil, fmt.Errorf("Incorrect json format: %s", err.Error())
	}

	return stringMaps, nil
}

func ParseJSONFromFileOrString(fileOrJSON string) (map[string]interface{}, error) {
	var jsonMap map[string]interface{}
	var err error
	var bytes []byte

	if fileOrJSON == "" {
		return nil, nil
	}

	if fileExists(fileOrJSON) {
		bytes, err = readJSONFile(fileOrJSON)
		if err != nil {
			return nil, err
		}
	} else {
		bytes = []byte(fileOrJSON)
	}

	jsonMap, err = parseJSON(bytes)

	if err != nil {
		return nil, err
	}

	return jsonMap, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readJSONFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func parseJSON(bytes []byte) (map[string]interface{}, error) {
	stringMap := map[string]interface{}{}
	err := json.Unmarshal(bytes, &stringMap)
	if err != nil {
		return nil, fmt.Errorf("Incorrect json format: %s", err.Error())
	}

	return stringMap, nil
}
