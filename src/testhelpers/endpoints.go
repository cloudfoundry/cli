package testhelpers

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
)

type RequestMatcher func(*http.Request) bool

var successMatcher = func(*http.Request) bool {
	return true
}

type TestResponse struct {
	Body string
	Status int
}

var RequestBodyMatcher = func(body string) RequestMatcher {
	return func(request *http.Request) bool {
		bodyBytes, err := ioutil.ReadAll(request.Body)

		if err != nil {
			return false
		}

		return string(bodyBytes) == body
	}
}

var CreateEndpoint = func(method string, path string, matcher RequestMatcher, response TestResponse) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if matcher == nil {
			matcher = successMatcher
		}

		acceptHeaderMatches := request.Header.Get("accept") == "application/json"
		authMatches := request.Header.Get("authorization") == "BEARER my_access_token"
		methodMatches := request.Method == method
		matcherMatches := matcher(request)

		paths := strings.Split(path, "?")
		pathMatches := request.URL.Path == paths[0]
		if len(paths) > 1 {
			queryStringMatches := strings.Contains(request.RequestURI, paths[1])
			pathMatches = pathMatches && queryStringMatches
		}

		if !(acceptHeaderMatches && authMatches && methodMatches && pathMatches && matcherMatches) {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(response.Status)
		fmt.Fprintln(writer, response.Body)
	}
}
