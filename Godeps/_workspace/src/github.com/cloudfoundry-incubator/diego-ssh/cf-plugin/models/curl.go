package models

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

type curlError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	ErrorCode   string `json:"error_code"`
}

type Curler func(cli plugin.CliConnection, result interface{}, args ...string) error

func Curl(cli plugin.CliConnection, result interface{}, args ...string) error {
	output, err := cli.CliCommandWithoutTerminalOutput(append([]string{"curl"}, args...)...)
	if err != nil {
		return err
	}

	buf := []byte(strings.Join(output, "\n"))

	var errorResponse curlError
	err = json.Unmarshal(buf, &errorResponse)
	if err != nil {
		return err
	}

	if errorResponse.Code != 0 {
		return errors.New(errorResponse.Description)
	}

	if result != nil {
		err = json.Unmarshal(buf, result)
		if err != nil {
			return err
		}
	}

	return nil
}
