package api

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
)

type RequestMatcher func (*http.Request) bool

var successMatcher = func(*http.Request) bool {
	return true
}

type TestResponse struct {
	Body   string
	Status int
	Header http.Header
}

func RemoveWhiteSpaceFromBody(body string) string {
	body = strings.Replace(body, " ", "", -1)
	body = strings.Replace(body, "\n", "", -1)
	body = strings.Replace(body, "\r", "", -1)
	body = strings.Replace(body, "\t", "", -1)
	return body
}

type RequestStatus struct {
	called bool
}

func (status *RequestStatus) Reset() {
	status.called = false
}

func (status *RequestStatus) Called() bool {
	return status.called
}

func (status *RequestStatus) call() {
	status.called = true
}

func matcherSequence(matchers []RequestMatcher) RequestMatcher {
	return func(request *http.Request) bool {
		for _, matcher := range matchers {
			if matcher != nil && !matcher(request) {
				return false
			}
		}
		return true
	}
}

func endpointCalledMatcher(status *RequestStatus) (matcher RequestMatcher) {
	status.Reset()
	matcher = func(*http.Request) bool {
		status.call()
		return true
	}
	return
}

func RequestBodyMatcher(expectedBody string) RequestMatcher {
	return func(request *http.Request) bool {
		bodyBytes, err := ioutil.ReadAll(request.Body)

		if err != nil {
			fmt.Printf("Error reading request body: %s", err.Error())
			return false
		}

		actualBody := string(bodyBytes)
		bodyMatches := actualBody == expectedBody

		if !bodyMatches {
			fmt.Printf("Body did not match. Expected [%s], Actual [%s]", expectedBody, actualBody)
		}
		return bodyMatches
	}
}

func CreateCheckableEndpoint(method string, path string, customMatcher RequestMatcher, response TestResponse) (hf http.HandlerFunc, status *RequestStatus) {
	status = &RequestStatus{}
	matchers := matcherSequence([]RequestMatcher{
		endpointCalledMatcher(status),
		customMatcher,
	})
	hf = createEndpoint(method, path , matchers, response)
	return
}

func createEndpoint(method string, path string, customMatcher RequestMatcher, response TestResponse) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		if customMatcher == nil {
			customMatcher = successMatcher
		}

		acceptHeaderMatches := request.Header.Get("accept") == "application/json"
		authMatches := strings.HasPrefix(request.Header.Get("authorization"), "BEARER my_access_token")
		methodMatches := request.Method == method
		customMatcherMatches := customMatcher(request)

		paths := strings.Split(path, "?")
		pathMatches := request.URL.Path == paths[0]
		if len(paths) > 1 {
			queryStringMatches := strings.Contains(request.RequestURI, paths[1])
			pathMatches = pathMatches && queryStringMatches
		}

		header := writer.Header()
		for name, values := range response.Header {
			if (len(values) < 1) {
				continue
			}
			header.Set(name, values[0])
		}

		if !(acceptHeaderMatches && authMatches && methodMatches && pathMatches && customMatcherMatches) {
			fmt.Printf("One of the matchers did not match. AcceptHeader [%t] Auth [%t] Method [%t] Path [%t] Custom Matcher [%t]",
				acceptHeaderMatches, authMatches, methodMatches, pathMatches, customMatcherMatches)

			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(response.Status)
		fmt.Fprintln(writer, response.Body)
	}
}
