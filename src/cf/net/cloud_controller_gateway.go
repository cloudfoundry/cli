package net

import (
	"cf/configuration"
	"encoding/json"
	"strconv"
)

type ccErrorResponse struct {
	Code        int
	Description string
}

var cloudControllerInvalidTokenCode = "1000"

func cloudControllerErrorHandler(body []byte) apiErrorInfo {
	ccResp := ccErrorResponse{}
	json.Unmarshal(body, &ccResp)

	code := strconv.Itoa(ccResp.Code)
	if code == cloudControllerInvalidTokenCode {
		code = INVALID_TOKEN_CODE
	}

	return apiErrorInfo{
		Code:        code,
		Description: ccResp.Description,
	}
}

func NewCloudControllerGateway(config configuration.Reader) Gateway {
	gateway := newGateway(cloudControllerErrorHandler, config)
	gateway.PollingEnabled = true
	return gateway
}
