package net

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

var uaaErrorHandler = func(statusCode int, body []byte) error {
	response := uaaErrorResponse{}
	json.Unmarshal(body, &response)

	if response.Code == "invalid_token" {
		return errors.NewInvalidTokenError(response.Description)
	}

	return errors.NewHttpError(statusCode, response.Code, response.Description)
}

func NewUAAGateway(config core_config.Reader, ui terminal.UI) Gateway {
	return newGateway(uaaErrorHandler, config, ui)
}
