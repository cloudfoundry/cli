package net

import (
	"encoding/json"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"strconv"
	"time"
)

type ccErrorResponse struct {
	Code        int
	Description string
}

func cloudControllerErrorHandler(statusCode int, body []byte) error {
	response := ccErrorResponse{}
	json.Unmarshal(body, &response)

	if response.Code == 1000 { // MAGICKAL NUMBERS AHOY
		return errors.NewInvalidTokenError(response.Description)
	} else {
		return errors.NewHttpError(statusCode, strconv.Itoa(response.Code), response.Description)
	}
}

func NewCloudControllerGateway(config core_config.Reader, clock func() time.Time) Gateway {
	gateway := newGateway(cloudControllerErrorHandler, config)
	gateway.Clock = clock
	gateway.PollingEnabled = true
	return gateway
}
