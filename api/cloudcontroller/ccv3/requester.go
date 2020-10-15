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

const MAX_QUERY_LENGTH = 100

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

type Requester interface {
	InitializeConnection(settings TargetSettings)

	InitializeRouter(resources map[string]string)

	MakeListRequest(requestParams RequestParams) (IncludedResources, Warnings, error)
	MakeListRequestWithPaginatedQuery(requestParams RequestParams) (IncludedResources, Warnings, error)

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

func (requester *RealRequester) MakeListRequest(requestParams RequestParams) (IncludedResources, Warnings, error) {
	request, err := requester.buildRequest(requestParams)
	if err != nil {
		return IncludedResources{}, nil, err
	}

	return requester.paginate(request, requestParams.ResponseBody, requestParams.AppendToList)
}

// note: this only paginates one query param
func queryParamToPaginate(requestParams RequestParams) (bool, int, []string) {
	for i, query := range requestParams.Query {
		if len(query.Values) > MAX_QUERY_LENGTH {
			return true, i, query.Values
		}
	}
	return false, 0, nil
}

func appendIncludedResources(includes IncludedResources, results IncludedResources) IncludedResources {
	includes.Users = append(includes.Users, results.Users...)
	includes.Organizations = append(includes.Organizations, results.Organizations...)
	includes.Spaces = append(includes.Spaces, results.Spaces...)
	includes.ServiceOfferings = append(includes.ServiceOfferings, results.ServiceOfferings...)
	includes.ServiceBrokers = append(includes.ServiceBrokers, results.ServiceBrokers...)
	return includes
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func (requester *RealRequester) MakeListRequestWithPaginatedQuery(requestParams RequestParams) (IncludedResources, Warnings, error) {
	var includes, warnings, batch = IncludedResources{}, Warnings{}, []string{}
	paginateQueryParams, idxToPaginate, queryParamVals := queryParamToPaginate(requestParams)
	if !paginateQueryParams {
		return requester.MakeListRequest(requestParams)
	}

	for len(queryParamVals) > 0 {
		nextBatchIdx := min(len(queryParamVals), MAX_QUERY_LENGTH)
		batch, queryParamVals = queryParamVals[:nextBatchIdx-1], queryParamVals[nextBatchIdx:]

		requestParams.Query[idxToPaginate].Values = batch
		results, listWarnings, err := requester.MakeListRequest(requestParams)
		if err != nil {
			return IncludedResources{}, nil, err
		}
		if warnings != nil {
			warnings = append(warnings, listWarnings...)
		}
		appendIncludedResources(includes, results)
	}
	return includes, warnings, nil
}

func (requester *RealRequester) MakeRequest(requestParams RequestParams) (JobURL, Warnings, error) {
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
