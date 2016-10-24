package net

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type JSONMapRequest map[string]interface{}

func (json *JSONMapRequest) String() string {
	return fmt.Sprintf("%#v", *json)
}

type JSONArrayRequest []interface{}

func (json *JSONArrayRequest) String() string {
	return fmt.Sprintf("%#v", *json)
}

func bytesToInterface(jsonBytes []byte) (interface{}, error) {
	mapResult := &JSONMapRequest{}
	err := json.Unmarshal(jsonBytes, mapResult)
	if err == nil {
		return mapResult, err
	}

	arrayResult := &JSONArrayRequest{}
	err = json.Unmarshal(jsonBytes, arrayResult)
	return arrayResult, err
}

func EmptyQueryParamMatcher() RequestMatcher {
	return func(request *http.Request) {
		defer GinkgoRecover()
		Expect(request.URL.RawQuery).To(Equal(""))
	}
}

func RequestBodyMatcher(expectedBodyString string) RequestMatcher {
	return func(request *http.Request) {
		defer GinkgoRecover()
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			Fail(fmt.Sprintf("Error reading request body: %s", err))
		}

		actualBody, err := bytesToInterface(bodyBytes)
		if err != nil {
			Fail(fmt.Sprintf("Error unmarshalling request: %s", err.Error()))
		}

		expectedBody, err := bytesToInterface([]byte(expectedBodyString))
		if err != nil {
			Fail(fmt.Sprintf("Error unmarshalling expected json: %s", err.Error()))
		}

		Expect(actualBody).To(Equal(expectedBody))
		Expect(request.Header.Get("content-type")).To(Equal("application/json"))
	}
}

func RequestBodyMatcherWithContentType(expectedBody, expectedContentType string) RequestMatcher {
	return func(request *http.Request) {
		defer GinkgoRecover()
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			Fail(fmt.Sprintf("Error reading request body: %s", err))
		}

		actualBody := string(bodyBytes)
		Expect(RemoveWhiteSpaceFromBody(actualBody)).To(Equal(RemoveWhiteSpaceFromBody(expectedBody)), "Body did not match.")

		actualContentType := request.Header.Get("content-type")
		Expect(actualContentType).To(Equal(expectedContentType), "Content Type did not match.")
	}
}

func RemoveWhiteSpaceFromBody(body string) string {
	body = strings.Replace(body, " ", "", -1)
	body = strings.Replace(body, "\n", "", -1)
	body = strings.Replace(body, "\r", "", -1)
	body = strings.Replace(body, "\t", "", -1)
	return body
}
