package net

import (
	"cf/configuration"
	"encoding/json"
)

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

var uaaInvalidTokenCode = "invalid_token"

var uaaErrorHandler = func(body []byte) apiErrorInfo {
	uaaResp := uaaErrorResponse{}
	json.Unmarshal(body, &uaaResp)

	code := uaaResp.Code
	if code == uaaInvalidTokenCode {
		code = INVALID_TOKEN_CODE
	}

	return apiErrorInfo{Code: code, Description: uaaResp.Description}
}

func NewUAAGateway(config configuration.Reader) Gateway {
	return newGateway(uaaErrorHandler, config)
}
