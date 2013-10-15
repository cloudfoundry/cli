package net

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func RequestBodyMatcher(expectedBody string) RequestMatcher {
	return RequestBodyMatcherWithContentType(expectedBody, "application/json")
}

func RequestBodyMatcherWithContentType(expectedBody, expectedContentType string) RequestMatcher {
	return func(request *http.Request) bool {
		bodyBytes, err := ioutil.ReadAll(request.Body)

		if err != nil {
			fmt.Printf("\nError reading request body: %s", err.Error())
			return false
		}

		actualBody := string(bodyBytes)
		bodyMatches := removeWhiteSpaceFromBody(actualBody) == removeWhiteSpaceFromBody(expectedBody)
		if !bodyMatches {
			fmt.Printf("\nBody did not match. Expected [%s], Actual [%s]", expectedBody, actualBody)
		}

		actualContentType := request.Header.Get("content-type")
		contentTypeMatches := actualContentType == expectedContentType
		if !contentTypeMatches {
			fmt.Printf("\nContent Type did not match. Expected [%s], Actual [%s]", expectedContentType, actualContentType)
		}

		return bodyMatches && contentTypeMatches
	}
}

func removeWhiteSpaceFromBody(body string) string {
	body = strings.Replace(body, " ", "", -1)
	body = strings.Replace(body, "\n", "", -1)
	body = strings.Replace(body, "\r", "", -1)
	body = strings.Replace(body, "\t", "", -1)
	return body
}
