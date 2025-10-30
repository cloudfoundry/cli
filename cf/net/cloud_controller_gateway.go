package net

import (
	"encoding/json"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/errors"
	"code.cloudfoundry.org/cli/v8/cf/terminal"
	"code.cloudfoundry.org/cli/v8/cf/trace"
)

type ccErrorResponse struct {
	Code        int
	Description string
}

type v3ErrorItem struct {
	Code   int
	Title  string
	Detail string
}

type v3CCError struct {
	Errors []v3ErrorItem
}

const invalidTokenCode = 1000

func cloudControllerErrorHandler(statusCode int, body []byte) error {
	response := ccErrorResponse{}
	_ = json.Unmarshal(body, &response)

	if response.Code == invalidTokenCode {
		return errors.NewInvalidTokenError(response.Description)
	}

	var v3response v3CCError
	_ = json.Unmarshal(body, &v3response)

	if len(v3response.Errors) > 0 && v3response.Errors[0].Code == invalidTokenCode {
		return errors.NewInvalidTokenError(v3response.Errors[0].Detail)
	}

	return errors.NewHTTPError(statusCode, strconv.Itoa(response.Code), response.Description)
}

func NewCloudControllerGateway(config coreconfig.Reader, clock func() time.Time, ui terminal.UI, logger trace.Printer, envDialTimeout string) Gateway {
	return Gateway{
		errHandler:      cloudControllerErrorHandler,
		config:          config,
		PollingThrottle: DefaultPollingThrottle,
		warnings:        &[]string{},
		Clock:           clock,
		ui:              ui,
		logger:          logger,
		PollingEnabled:  true,
		DialTimeout:     dialTimeout(envDialTimeout),
	}
}
