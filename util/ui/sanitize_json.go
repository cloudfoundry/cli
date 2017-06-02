package ui

import (
	"bytes"
	"encoding/json"
	"regexp"
)

var keysToSanitize = regexp.MustCompile("(?i).*(?:token|password).*")

const tokenEndpoint = "token_endpoint"

func SanitizeJSON(raw []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	decoder := json.NewDecoder(bytes.NewBuffer(raw))
	decoder.UseNumber()
	err := decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	return iterateAndRedact(result), nil
}

func iterateAndRedact(blob map[string]interface{}) map[string]interface{} {
	for key, value := range blob {
		switch v := value.(type) {
		case string:
			if keysToSanitize.Match([]byte(key)) && key != tokenEndpoint {
				blob[key] = RedactedValue
			}
		case map[string]interface{}:
			blob[key] = iterateAndRedact(v)
		}
	}

	return blob
}
