package ccv3

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"encoding/json"
)

func (client *Client) makeCreateRequest(requestName string, requestBody interface{}, responseBody interface{}) (Warnings, error) {
	return client.makeCreateRequestWithOptions(
		requestOptions{
			RequestName: requestName,
		},
		requestBody,
		responseBody,
	)
}

func (client *Client) makeCreateRequestWithParams(requestName string, uriParams internal.Params, requestBody interface{}, responseBody interface{}) (Warnings, error) {
	return client.makeCreateRequestWithOptions(
		requestOptions{
			RequestName: requestName,
			URIParams:   uriParams,
		},
		requestBody,
		responseBody,
	)
}

func (client *Client) makeCreateRequestWithOptions(requestOptions requestOptions, requestBody interface{}, responseBody interface{}) (Warnings, error) {
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	requestOptions.Body = bytes.NewReader(body)
	request, err := client.newHTTPRequest(requestOptions)
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{DecodeJSONResponseInto: &responseBody}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
