package util

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func GetJSONFromFlagValue(input string) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, fmt.Errorf("invalid input: %s", input)
	}

	trimmedInput := strings.Trim(input, `"'`)
	if strings.HasPrefix(trimmedInput, `@`) {
		trimmedInput = strings.Trim(trimmedInput[1:], `"'`)
		jsonBytes, err := ioutil.ReadFile(trimmedInput)
		if err != nil {
			return []byte{}, err
		}

		return jsonBytes, nil
	}

	jsonBytes, err := ioutil.ReadFile(trimmedInput)
	if err != nil {
		if strings.ContainsAny(trimmedInput, "[{") {
			return []byte(trimmedInput), nil
		}
		return []byte{}, err
	}

	return jsonBytes, nil
}
