package net

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type errorResponse struct {
	Name    string
	Message string
}

func errorHandler(statusCode int, body []byte) error {
	response := errorResponse{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		return errors.NewHttpError(http.StatusInternalServerError, "", "")
	}

	return errors.NewHttpError(statusCode, response.Name, response.Message)
}

func NewRoutingApiGateway(config core_config.Reader, clock func() time.Time, ui terminal.UI) Gateway {
	gateway := newGateway(errorHandler, config, ui)
	gateway.Clock = clock
	gateway.PollingEnabled = true
	return gateway
}
