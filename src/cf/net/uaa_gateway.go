package net

import (
	"cf/configuration"
	"cf/errors"
	"encoding/json"
)

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

var uaaErrorHandler = func(statusCode int, body []byte) error {
	uaaResp := uaaErrorResponse{}
	json.Unmarshal(body, &uaaResp)

	if uaaResp.Code == "invalid_token" {
		return errors.NewInvalidTokenError(uaaResp.Description)
	} else {
		return errors.NewHttpError(statusCode, uaaResp.Code, uaaResp.Description)
	}
}

func NewUAAGateway(config configuration.Reader) Gateway {
	return newGateway(uaaErrorHandler, config)
}
