package net

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
)

type ccErrorResponse struct {
	Code        int
	Description string
}

const invalidTokenCode = 1000

func cloudControllerErrorHandler(statusCode int, body []byte) error {
	response := ccErrorResponse{}
	json.Unmarshal(body, &response)

	if response.Code == invalidTokenCode {
		return errors.NewInvalidTokenError(response.Description)
	}

	return errors.NewHTTPError(statusCode, strconv.Itoa(response.Code), response.Description)
}

func NewCloudControllerGateway(config coreconfig.Reader, clock func() time.Time, ui terminal.UI, logger trace.Printer) Gateway {
	gateway := newGateway(cloudControllerErrorHandler, config, ui, logger)
	gateway.Clock = clock
	gateway.PollingEnabled = true
	return gateway
}
