package net

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"errors"
)

func RequestBodyMatcher(expectedBody string) RequestMatcher {
	return RequestBodyMatcherWithContentType(expectedBody, "application/json")
}

func RequestBodyMatcherWithContentType(expectedBody, expectedContentType string) RequestMatcher {
	return func(request *http.Request) error {
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return err
		}

		actualBody := string(bodyBytes)
		bodyMatches := removeWhiteSpaceFromBody(actualBody) == removeWhiteSpaceFromBody(expectedBody)
		if !bodyMatches {
			return errors.New(fmt.Sprintf("\nBody did not match. Expected [%s], Actual [%s]", expectedBody, actualBody))
		}

		actualContentType := request.Header.Get("content-type")
		contentTypeMatches := actualContentType == expectedContentType
		if !contentTypeMatches {
			return errors.New(fmt.Sprintf("\nContent Type did not match. Expected [%s], Actual [%s]", expectedContentType, actualContentType))
		}

		return nil
	}
}

func removeWhiteSpaceFromBody(body string) string {
	body = strings.Replace(body, " ", "", -1)
	body = strings.Replace(body, "\n", "", -1)
	body = strings.Replace(body, "\r", "", -1)
	body = strings.Replace(body, "\t", "", -1)
	return body
}
