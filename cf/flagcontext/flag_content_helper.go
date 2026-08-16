package flagcontext

import (
	"fmt"
	"os"
	"strings"
)

func GetContentsFromFlagValue(input string) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, fmt.Errorf("invalid input: %s", input)
	}

	return GetContentsFromOptionalFlagValue(input)
}

func GetContentsFromOptionalFlagValue(input string) ([]byte, error) {
	trimmedInput := strings.Trim(input, `"'`)
	if strings.HasPrefix(trimmedInput, `@`) {
		trimmedInput = strings.Trim(trimmedInput[1:], `"'`)
		bs, err := os.ReadFile(trimmedInput)
		if err != nil {
			return []byte{}, err
		}

		return bs, nil
	}

	bs, err := os.ReadFile(trimmedInput)
	if err != nil {
		return []byte(trimmedInput), nil
	}

	return bs, nil
}
