package net

import (
	"cf/configuration"
	"cf/errors"
	"encoding/json"
	"strconv"
)

type ccErrorResponse struct {
	Code        int
	Description string
}

func cloudControllerErrorHandler(statusCode int, body []byte) error {
	ccResp := ccErrorResponse{}
	json.Unmarshal(body, &ccResp)
	code := strconv.Itoa(ccResp.Code)

	if code == "1000" {
		return errors.NewInvalidTokenError(ccResp.Description)
	} else {
		return errors.NewHttpError(statusCode, code, ccResp.Description)
	}
}

func NewCloudControllerGateway(config configuration.Reader) Gateway {
	gateway := newGateway(cloudControllerErrorHandler, config)
	gateway.PollingEnabled = true
	return gateway
}
