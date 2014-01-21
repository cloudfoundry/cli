package net

import "net/http"

func NewCurlGateway() Gateway {
	errorHandler := func(response *http.Response) errorResponse {
		return errorResponse{}
	}
	gateway := newGateway(errorHandler)
	gateway.PollingEnabled = true
	return gateway
}
