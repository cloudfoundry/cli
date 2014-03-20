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
	response := ccErrorResponse{}
	json.Unmarshal(body, &response)

	if response.Code == 1000 {
		return errors.NewInvalidTokenError(response.Description)
	} else {
		return errors.NewHttpError(statusCode, strconv.Itoa(response.Code), response.Description)
	}
}

func NewCloudControllerGateway(config configuration.Reader) Gateway {
	gateway := newGateway(cloudControllerErrorHandler, config)
	gateway.PollingEnabled = true
	return gateway
}
