package net

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
	"encoding/json"
	"fmt"
)

type JSONMapRequest map[string]interface{}
func (json *JSONMapRequest) String() string {
	return fmt.Sprintf("%#v",*json)
}

type JSONArrayRequest []interface{}
func (json *JSONArrayRequest) String() string {
	return fmt.Sprintf("%#v",*json)
}

func bytesToInterface(jsonBytes []byte) (interface {}, error){
	mapResult := &JSONMapRequest{}
	err := json.Unmarshal(jsonBytes,mapResult)
	if err == nil {
		return mapResult, err
	}

	arrayResult := &JSONArrayRequest{}
	err = json.Unmarshal(jsonBytes,arrayResult)
	return arrayResult, err
}

func RequestBodyMatcher(expectedBodyString string) RequestMatcher {
	return func(t *testing.T, request *http.Request) {
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			assert.Fail(t,"Error reading request body: %s",err)
		}

		actualBody, err := bytesToInterface(bodyBytes)
		if err != nil {
			assert.Fail(t,"Error unmarshalling request",err.Error())
		}

		expectedBody, err :=  bytesToInterface([]byte(expectedBodyString))
		if err != nil {
			assert.Fail(t,"Error unmarshalling expected json",err.Error())
		}

		assert.Equal(t,expectedBody,actualBody,"\nEXPECTED: %s\nACTUAL:   %s",expectedBody,actualBody)
		assert.Equal(t,request.Header.Get("content-type"), "application/json", "Content Type was not application/json.")
	}
}

func RequestBodyMatcherWithContentType(expectedBody, expectedContentType string) RequestMatcher {
	return func(t *testing.T, request *http.Request) {
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			assert.Fail(t,"Error reading request body: %s",err)
		}

		actualBody := string(bodyBytes)
		assert.Equal(t,RemoveWhiteSpaceFromBody(actualBody),RemoveWhiteSpaceFromBody(expectedBody), "Body did not match.")

		actualContentType := request.Header.Get("content-type")
		assert.Equal(t,actualContentType,expectedContentType, "Content Type did not match.")
	}
}

func RemoveWhiteSpaceFromBody(body string) string {
	body = strings.Replace(body, " ", "", -1)
	body = strings.Replace(body, "\n", "", -1)
	body = strings.Replace(body, "\r", "", -1)
	body = strings.Replace(body, "\t", "", -1)
	return body
}
