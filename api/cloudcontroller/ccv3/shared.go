package ccv3

import (
	"bytes"
	"encoding/json"
	"io"

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

func (client *Client) MakeRequestReceiveRaw(
	requestName string,
	uriParams internal.Params,
	responseBodyMimeType string,
) ([]byte, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   uriParams,
	})
	if err != nil {
		return nil, nil, err
	}

	response := cloudcontroller.Response{}

	request.Header.Set("Accept", responseBodyMimeType)

	err = client.connection.Make(request, &response)

	return response.RawResponse, response.Warnings, err
}

func (client *Client) MakeRequestSendRaw(
	requestName string,
	uriParams internal.Params,
	requestBody []byte,
	requestBodyMimeType string,
	responseBody interface{},
) (string, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   uriParams,
		Body:        bytes.NewReader(requestBody),
	})
	if err != nil {
		return "", nil, err
	}

	request.Header.Set("Content-type", requestBodyMimeType)

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: responseBody,
	}

	err = client.connection.Make(request, &response)

	return response.ResourceLocationURL, response.Warnings, err
}

func (client *Client) MakeRequestUploadAsync(
	requestName string,
	uriParams internal.Params,
	requestBodyMimeType string,
	requestBody io.ReadSeeker,
	dataLength int64,
	responseBody interface{},
	writeErrors <-chan error,
) (string, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   uriParams,
		Body:        requestBody,
	})
	if err != nil {
		return "", nil, err
	}

	request.Header.Set("Content-Type", requestBodyMimeType)
	request.ContentLength = dataLength

	return client.uploadAsynchronously(request, responseBody, writeErrors)
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

func (client *Client) uploadAsynchronously(request *cloudcontroller.Request, responseBody interface{}, writeErrors <-chan error) (string, Warnings, error) {
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: responseBody,
	}

	httpErrors := make(chan error)

	go func() {
		defer close(httpErrors)

		err := client.connection.Make(request, &response)
		if err != nil {
			httpErrors <- err
		}
	}()

	// The following section makes the following assumptions:
	// 1) If an error occurs during file reading, an EOF is sent to the request
	// object. Thus ending the request transfer.
	// 2) If an error occurs during request transfer, an EOF is sent to the pipe.
	// Thus ending the writing routine.
	var firstError error
	var writeClosed, httpClosed bool

	for {
		select {
		case writeErr, ok := <-writeErrors:
			if !ok {
				writeClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = writeErr
			}
		case httpErr, ok := <-httpErrors:
			if !ok {
				httpClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = httpErr
			}
		}

		if writeClosed && httpClosed {
			break // for for
		}
	}

	return response.ResourceLocationURL, response.Warnings, firstError
}
