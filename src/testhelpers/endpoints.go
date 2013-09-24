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

func RemoveWhiteSpaceFromBody(body string) string {
	body = strings.Replace(body, " ", "", -1)
	body = strings.Replace(body, "\n", "", -1)
	body = strings.Replace(body, "\r", "", -1)
	body = strings.Replace(body, "\t", "", -1)
	return body
}

var RequestBodyMatcher = func(expectedBody string) RequestMatcher {
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

var CreateEndpoint = func(method string, path string, customMatcher RequestMatcher, response TestResponse) http.HandlerFunc {
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
