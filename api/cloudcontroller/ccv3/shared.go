package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type RequestParams struct {
	RequestName    string
	URIParams      internal.Params
	Query          []Query
	RequestBody    interface{}
	RequestHeaders [][]string
	ResponseBody   interface{}
	URL            string
	AppendToList   func(item interface{}) error
}

type RequestParamsForSendRaw struct {
	RequestName         string
	URIParams           internal.Params
	Query               []Query
	RequestBody         []byte
	RequestHeaders      [][]string
	ResponseBody        interface{}
	URL                 string
	RequestBodyMimeType string
}

type RequestParamsForReceiveRaw struct {
	RequestName          string
	URIParams            internal.Params
	Query                []Query
	RequestHeaders       [][]string
	ResponseBody         interface{}
	URL                  string
	ResponseBodyMimeType string
}

func (client *Client) MakeListRequest(requestParams RequestParams) (IncludedResources, Warnings, error) {
	request, err := client.buildRequest(requestParams)
	if err != nil {
		return IncludedResources{}, nil, err
	}

	return client.paginate(request, requestParams.ResponseBody, requestParams.AppendToList)
}

func (client *Client) MakeRequest(requestParams RequestParams) (JobURL, Warnings, error) {
	request, err := client.buildRequest(requestParams)
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = requestParams.ResponseBody
	}

	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

func (client *Client) MakeRequestReceiveRaw(requestParams RequestParamsForReceiveRaw) ([]byte, Warnings, error) {
	options := requestOptions{
		RequestName: requestParams.RequestName,
		URIParams:   requestParams.URIParams,
		Query:       requestParams.Query,
		URL:         requestParams.URL,
	}
	request, err := client.newHTTPRequest(options)
	if err != nil {
		return nil, nil, err
	}

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = requestParams.ResponseBody
	}

	request.Header.Set("Accept", requestParams.ResponseBodyMimeType)

	err = client.connection.Make(request, &response)

	return response.RawResponse, response.Warnings, err
}

func (client *Client) MakeRequestSendRaw(requestParams RequestParamsForSendRaw) (string, Warnings, error) {
	options := requestOptions{
		RequestName: requestParams.RequestName,
		URIParams:   requestParams.URIParams,
		Query:       requestParams.Query,
		URL:         requestParams.URL,
		Body:        bytes.NewReader(requestParams.RequestBody),
	}

	request, err := client.newHTTPRequest(options)
	if err != nil {
		return "", nil, err
	}

	request.Header.Set("Content-type", requestParams.RequestBodyMimeType)

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = requestParams.ResponseBody
	}

	err = client.connection.Make(request, &response)

	return response.ResourceLocationURL, response.Warnings, err
}

func (client *Client) buildRequest(requestParams RequestParams) (*cloudcontroller.Request, error) {
	options := requestOptions{
		RequestName: requestParams.RequestName,
		URIParams:   requestParams.URIParams,
		Query:       requestParams.Query,
		URL:         requestParams.URL,
	}

	if requestParams.RequestBody != nil {
		body, err := json.Marshal(requestParams.RequestBody)
		if err != nil {
			return nil, err
		}

		options.Body = bytes.NewReader(body)
	}

	request, err := client.newHTTPRequest(options)
	if err != nil {
		return nil, err
	}

	return request, err
}
