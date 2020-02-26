package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

//go:generate counterfeiter . Requester

type Requester interface {
	InitializeConnection(settings TargetSettings)

	InitializeRouter(resources map[string]string)

	MakeListRequest(requestParams RequestParams) (IncludedResources, Warnings, error)

	MakeRequest(requestParams RequestParams) (JobURL, Warnings, error)

	MakeRequestReceiveRaw(
		requestName string,
		uriParams internal.Params,
		responseBodyMimeType string,
	) ([]byte, Warnings, error)

	MakeRequestSendRaw(
		requestName string,
		uriParams internal.Params,
		requestBody []byte,
		requestBodyMimeType string,
		responseBody interface{},
	) (string, Warnings, error)

	MakeRequestUploadAsync(
		requestName string,
		uriParams internal.Params,
		requestBodyMimeType string,
		requestBody io.ReadSeeker,
		dataLength int64,
		responseBody interface{},
		writeErrors <-chan error,
	) (string, Warnings, error)

	WrapConnection(wrapper ConnectionWrapper)
}

type RealRequester struct {
	connection cloudcontroller.Connection
	router     *internal.Router
	userAgent  string
	wrappers   []ConnectionWrapper
}

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

func NewRequester(config Config) *RealRequester {
	userAgent := fmt.Sprintf(
		"%s/%s (%s; %s %s)",
		config.AppName,
		config.AppVersion,
		runtime.Version(),
		runtime.GOARCH,
		runtime.GOOS,
	)

	return &RealRequester{
		userAgent: userAgent,
		wrappers:  append([]ConnectionWrapper{newErrorWrapper()}, config.Wrappers...),
	}
}

func (requester *RealRequester) InitializeConnection(settings TargetSettings) {
	requester.connection = cloudcontroller.NewConnection(cloudcontroller.Config{
		DialTimeout:       settings.DialTimeout,
		SkipSSLValidation: settings.SkipSSLValidation,
	})

	for _, wrapper := range requester.wrappers {
		requester.connection = wrapper.Wrap(requester.connection)
	}
}

func (requester *RealRequester) InitializeRouter(resources map[string]string) {
	requester.router = internal.NewRouter(internal.APIRoutes, resources)
}

func (requester *RealRequester) MakeListRequest(
	requestParams RequestParams,
) (IncludedResources, Warnings, error) {
	request, err := requester.buildRequest(requestParams)
	if err != nil {
		return IncludedResources{}, nil, err
	}

	return requester.paginate(request, requestParams.ResponseBody, requestParams.AppendToList)
}

func (requester *RealRequester) MakeRequest(
	requestParams RequestParams,
) (JobURL, Warnings, error) {
	request, err := requester.buildRequest(requestParams)
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	if requestParams.ResponseBody != nil {
		response.DecodeJSONResponseInto = requestParams.ResponseBody
	}

	err = requester.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

func (requester *RealRequester) MakeRequestReceiveRaw(
	requestName string,
	uriParams internal.Params,
	responseBodyMimeType string,
) ([]byte, Warnings, error) {
	request, err := requester.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   uriParams,
	})
	if err != nil {
		return nil, nil, err
	}

	response := cloudcontroller.Response{}

	request.Header.Set("Accept", responseBodyMimeType)

	err = requester.connection.Make(request, &response)

	return response.RawResponse, response.Warnings, err
}

func (requester *RealRequester) MakeRequestSendRaw(
	requestName string,
	uriParams internal.Params,
	requestBody []byte,
	requestBodyMimeType string,
	responseBody interface{},
) (string, Warnings, error) {
	request, err := requester.newHTTPRequest(requestOptions{
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

	err = requester.connection.Make(request, &response)

	return response.ResourceLocationURL, response.Warnings, err
}

func (requester *RealRequester) MakeRequestUploadAsync(
	requestName string,
	uriParams internal.Params,
	requestBodyMimeType string,
	requestBody io.ReadSeeker,
	dataLength int64,
	responseBody interface{},
	writeErrors <-chan error,
) (string, Warnings, error) {
	request, err := requester.newHTTPRequest(requestOptions{
		RequestName: requestName,
		URIParams:   uriParams,
		Body:        requestBody,
	})
	if err != nil {
		return "", nil, err
	}

	request.Header.Set("Content-Type", requestBodyMimeType)
	request.ContentLength = dataLength

	return requester.uploadAsynchronously(request, responseBody, writeErrors)
}

func (requester *RealRequester) buildRequest(requestParams RequestParams) (*cloudcontroller.Request, error) {
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

	request, err := requester.newHTTPRequest(options)
	if err != nil {
		return nil, err
	}

	return request, err
}

func (requester *RealRequester) uploadAsynchronously(request *cloudcontroller.Request, responseBody interface{}, writeErrors <-chan error) (string, Warnings, error) {
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: responseBody,
	}

	httpErrors := make(chan error)

	go func() {
		defer close(httpErrors)

		err := requester.connection.Make(request, &response)
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
