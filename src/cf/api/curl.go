package api

import (
	"bufio"
	"cf/configuration"
	"cf/net"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strings"
)

type CurlRepository interface {
	Request(method, path, header, body string) (resHeaders, resBody string, apiResponse net.ApiResponse)
}

type CloudControllerCurlRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerCurlRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerCurlRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerCurlRepository) Request(method, path, headerString, body string) (resHeaders, resBody string, apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/%s", repo.config.ApiEndpoint(), strings.TrimLeft(path, "/"))

	req, apiResponse := repo.gateway.NewRequest(method, url, repo.config.AccessToken(), strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	err := mergeHeaders(req.HttpReq.Header, headerString)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error parsing headers", err)
		return
	}

	res, apiResponse := repo.gateway.PerformRequestForResponse(req)

	if apiResponse.IsNotSuccessful() {
		if apiResponse.IsHttpError() {
			resHeaders = apiResponse.ErrorHeader
			resBody = apiResponse.ErrorBody
			apiResponse = net.NewSuccessfulApiResponse()
			return
		} else {
			return
		}
	}

	headerBytes, _ := httputil.DumpResponse(res, false)
	resHeaders = string(headerBytes)

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error reading response", err)
	}
	resBody = string(bytes)

	return
}

func mergeHeaders(destination http.Header, headerString string) (err error) {
	headerString = strings.TrimSpace(headerString)
	headerString += "\n\n"
	headerReader := bufio.NewReader(strings.NewReader(headerString))
	headers, err := textproto.NewReader(headerReader).ReadMIMEHeader()
	if err != nil {
		return
	}

	for key, values := range headers {
		destination.Del(key)
		for _, value := range values {
			destination.Add(key, value)
		}
	}

	return
}
