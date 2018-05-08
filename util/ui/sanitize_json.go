package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

// tokenEndpoint is explicitly excluded from sanitization
const tokenEndpoint = "token_endpoint"

var keysToSanitize = regexp.MustCompile("(?i)token|password")
var sanitizeValues = regexp.MustCompile(`([&?]password)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)

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
			if keysToSanitize.MatchString(key) && key != tokenEndpoint {
				blob[key] = RedactedValue
			} else {
				blob[key] = sanitizeValues.ReplaceAllString(value.(string), fmt.Sprintf("$1=%s", RedactedValue))
			}
		case map[string]interface{}:
			blob[key] = iterateAndRedact(v)
		}
	}

	return blob
}
