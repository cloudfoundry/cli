package net

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

var uaaErrorHandler = func(response *http.Response) errorResponse {
	invalidTokenCode := "invalid_token"

	jsonBytes, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	uaaResp := uaaErrorResponse{}
	json.Unmarshal(jsonBytes, &uaaResp)

	code := uaaResp.Code
	if code == invalidTokenCode {
		code = INVALID_TOKEN_CODE
	}

	return errorResponse{Code: code, Description: uaaResp.Description}
}

func NewUAAGateway() Gateway {
	return newGateway(uaaErrorHandler)
}
