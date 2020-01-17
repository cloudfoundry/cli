package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type requestParams struct {
	RequestName  string
	URIParams    internal.Params
	Query        []Query
	RequestBody  interface{}
	ResponseBody interface{}
}

func (client *Client) makeRequest(requestParams requestParams) (JobURL, Warnings, error) {
	options := requestOptions{
		RequestName: requestParams.RequestName,
		URIParams:   requestParams.URIParams,
		Query:       requestParams.Query,
	}

	if requestParams.RequestBody != nil {
		body, err := json.Marshal(requestParams.RequestBody)
		if err != nil {
			return "", nil, err
		}

		options.Body = bytes.NewReader(body)
	}

	request, err := client.newHTTPRequest(options)
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = &requestParams.ResponseBody
	}

	err = client.connection.Make(request, &response)

	return client.GetSingleResponse(requestParams, request)
}

func (client *Client) makeListRequest(requestParams requestParams) ([]interface{}, Warnings, error) {
	options := requestOptions{
		RequestName: requestParams.RequestName,
		URIParams:   requestParams.URIParams,
		Query:       requestParams.Query,
	}

	if requestParams.RequestBody != nil {
		body, err := json.Marshal(requestParams.RequestBody)
		if err != nil {
			return nil, nil, err
		}

		options.Body = bytes.NewReader(body)
	}

	request, err := client.newHTTPRequest(options)
	if err != nil {
		return nil, nil, err
	}

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = &requestParams.ResponseBody
	}

	err = client.connection.Make(request, &response)

	return client.GetListResponse(requestParams, request)
}

func (client *Client) GetSingleResponse(requestParams requestParams, request *cloudcontroller.Request) (JobURL, Warnings, error) {
	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = &requestParams.ResponseBody
	}

	err := client.connection.Make(request, &response)
	// unmarshals object of correct type into &response using unsafe.pointer directly to actor

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

func (client *Client) GetListResponse(requestParams requestParams, request *cloudcontroller.Request) ([]interface{}, Warnings, error) {
	var fullResourceList []interface{}

	warnings, err := client.paginate(request, requestParams.ResponseBody, func(item interface{}) error {
		fullResourceList = append(fullResourceList, item)
		return nil
	})

	return fullResourceList, warnings, err
}
