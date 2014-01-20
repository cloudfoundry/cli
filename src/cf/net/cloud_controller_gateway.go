package net

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

func NewCloudControllerGateway() Gateway {
	invalidTokenCode := "1000"

	type ccErrorResponse struct {
		Code        int
		Description string
	}

	errorHandler := func(response *http.Response) ErrorResponse {
		jsonBytes, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()

		ccResp := ccErrorResponse{}
		json.Unmarshal(jsonBytes, &ccResp)

		code := strconv.Itoa(ccResp.Code)
		if code == invalidTokenCode {
			code = INVALID_TOKEN_CODE
		}

		return ErrorResponse{Code: code, Description: ccResp.Description}
	}

	gateway := newGateway(errorHandler)
	gateway.PollingEnabled = true
	return gateway
}
