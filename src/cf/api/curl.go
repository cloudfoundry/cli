package api

import (
	"bufio"
	"cf/configuration"
	"cf/errors"
	"cf/net"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strings"
)

type CurlRepository interface {
	Request(method, path, header, body string) (resHeaders, resBody string, apiErr error)
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

func (repo CloudControllerCurlRepository) Request(method, path, headerString, body string) (resHeaders, resBody string, apiErr error) {
	url := fmt.Sprintf("%s/%s", repo.config.ApiEndpoint(), strings.TrimLeft(path, "/"))

	req, apiErr := repo.gateway.NewRequest(method, url, repo.config.AccessToken(), strings.NewReader(body))
	if apiErr != nil {
		return
	}

	err := mergeHeaders(req.HttpReq.Header, headerString)
	if err != nil {
		apiErr = errors.NewErrorWithError("Error parsing headers", err)
		return
	}

	res, apiErr := repo.gateway.PerformRequestForResponse(req)

	if apiErr != nil {
		if httpErr, ok := apiErr.(errors.HttpError); ok {
			resHeaders = httpErr.Headers()
			resBody = httpErr.Body()
			apiErr = nil
		}

		return
	}

	headerBytes, _ := httputil.DumpResponse(res, false)
	resHeaders = string(headerBytes)

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		apiErr = errors.NewErrorWithError("Error reading response", err)
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
