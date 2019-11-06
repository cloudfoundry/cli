package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

var keysToSanitize = regexp.MustCompile("(?i)token|password")
var sanitizeURIParams = regexp.MustCompile(`([&?]password)=[A-Za-z0-9\-._~!$'()*+,;=:@/?]*`)
var sanitizeURLPassword = regexp.MustCompile(`([\d\w]+):\/\/([^:]+):(?:[^@]+)@`)

func SanitizeJSON(raw []byte) ([]byte, error) {
	var result interface{}
	decoder := json.NewDecoder(bytes.NewBuffer(raw))
	decoder.UseNumber()
	err := decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	sanitized := iterateAndRedact(result)

	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(sanitized)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func iterateAndRedact(blob interface{}) interface{} {
	switch v := blob.(type) {
	case string:
		return sanitizeURL(v)
	case []interface{}:
		var list []interface{}
		for _, val := range v {
			list = append(list, iterateAndRedact(val))
		}

		return list
	case map[string]interface{}:
		for key, value := range v {
			if keysToSanitize.MatchString(key) {
				v[key] = RedactedValue
			} else {
				v[key] = iterateAndRedact(value)
			}
		}
		return v
	}
	return blob
}

func sanitizeURL(rawURL string) string {
	sanitized := sanitizeURLPassword.ReplaceAllString(rawURL, fmt.Sprintf("$1://$2:%s@", RedactedValue))
	sanitized = sanitizeURIParams.ReplaceAllString(sanitized, fmt.Sprintf("$1=%s", RedactedValue))
	return sanitized
}
