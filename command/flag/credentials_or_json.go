package flag

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type CredentialsOrJSON struct {
	IsSet                 bool
	UserPromptCredentials []string
	Value                 map[string]interface{}
}

func (c *CredentialsOrJSON) UnmarshalFlag(input string) error {
	c.IsSet = true

	input = strings.Trim(input, `"'`)

	if value, ok := canParseAsJSON(input); ok {
		c.Value = value
		return nil
	}

	value, ok, err := canParseFileAsJSON(input)
	if err != nil {
		return err
	}
	if ok {
		c.Value = value
		return nil
	}

	keys, ok := canParseAsCredentialsKeys(input)
	if ok {
		c.UserPromptCredentials = keys
	}

	return nil
}

func (c *CredentialsOrJSON) Complete(prefix string) []flags.Completion {
	return completeWithTilde(prefix)
}

func canParseAsJSON(input string) (value map[string]interface{}, ok bool) {
	if err := json.Unmarshal([]byte(input), &value); err == nil {
		ok = true
	}
	return
}

func canParseFileAsJSON(input string) (value map[string]interface{}, ok bool, parseError error) {
	contents, err := ioutil.ReadFile(input)
	if err != nil {
		return
	}

	if err := json.Unmarshal(contents, &value); err != nil {
		parseError = &flags.Error{
			Type:    flags.ErrRequired,
			Message: fmt.Sprintf("The file '%s' contains invalid JSON. Please provide a path to a file containing a valid JSON object.", input),
		}
		return
	}

	ok = true
	return
}

func canParseAsCredentialsKeys(input string) (credentials []string, ok bool) {
	if len(input) > 0 {
		ok = true
		for _, key := range strings.Split(input, ",") {
			key = strings.Trim(key, " ")
			credentials = append(credentials, key)
		}
	}
	return
}
