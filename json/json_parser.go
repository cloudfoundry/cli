package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ParseJsonArray(path string) ([]map[string]interface{}, error) {
	if path == "" {
		return nil, nil
	}

	bytes, err := readJsonFile(path)
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

func ParseJsonFromFileOrString(fileOrJson string) (map[string]interface{}, error) {
	var jsonMap map[string]interface{}
	var err error
	var bytes []byte

	if fileOrJson == "" {
		return nil, nil
	}

	if fileExists(fileOrJson) {
		bytes, err = readJsonFile(fileOrJson)
		if err != nil {
			return nil, err
		}
	} else {
		bytes = []byte(fileOrJson)
	}

	jsonMap, err = parseJson(bytes)

	if err != nil {
		return nil, err
	}

	return jsonMap, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readJsonFile(path string) ([]byte, error) {
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

func parseJson(bytes []byte) (map[string]interface{}, error) {
	stringMap := map[string]interface{}{}
	err := json.Unmarshal(bytes, &stringMap)
	if err != nil {
		return nil, fmt.Errorf("Incorrect json format: %s", err.Error())
	}

	return stringMap, nil
}
