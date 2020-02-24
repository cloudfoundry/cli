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
	RawBody        []byte
	RequestBody    interface{}
	RequestHeaders [][]string
	ResponseBody   interface{}
	URL            string
	AppendToList   func(item interface{}) error
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

func (client *Client) MakeReceiveRawRequest(requestParams RequestParams) ([]byte, Warnings, error) {
	request, err := client.buildRequest(requestParams)
	if err != nil {
		return nil, nil, err
	}

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = requestParams.ResponseBody
	}

	err = client.connection.Make(request, &response)

	return response.RawResponse, response.Warnings, err
}

func (client *Client) MakeSendRawRequest(requestParams RequestParams) (JobURL, Warnings, error) {
	options := requestOptions{
		RequestName: requestParams.RequestName,
		URIParams:   requestParams.URIParams,
		Query:       requestParams.Query,
		URL:         requestParams.URL,
		Body:        bytes.NewReader(requestParams.RawBody),
	}

	request, err := client.newHTTPRequest(options)
	if err != nil {
		return "", nil, err
	}

	if requestParams.AcceptMimeType != "" {
		request.Header.Set("Accept", requestParams.AcceptMimeType)
	}

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = requestParams.ResponseBody
	}

	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
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
	} else if requestParams.RawBody != nil {
		options.Body = bytes.NewReader(requestParams.RawBody)
	}

	request, err := client.newHTTPRequest(options)
	if err != nil {
		return nil, err
	}

	if requestParams.RequestHeaders != nil {
		for _, header := range requestParams.RequestHeaders {
			request.Header.Add(header[0], header[1])
		}
	}

	return request, err
}
