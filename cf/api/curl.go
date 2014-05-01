package api

import (
	"bufio"
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/net"
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

func (repo CloudControllerCurlRepository) Request(method, path, headerString, body string) (resHeaders, resBody string, err error) {
	url := fmt.Sprintf("%s/%s", repo.config.ApiEndpoint(), strings.TrimLeft(path, "/"))

	req, err := repo.gateway.NewRequest(method, url, repo.config.AccessToken(), strings.NewReader(body))
	if err != nil {
		return
	}

	err = mergeHeaders(req.HttpReq.Header, headerString)
	if err != nil {
		err = errors.NewWithError("Error parsing headers", err)
		return
	}

	res, err := repo.gateway.PerformRequest(req)

	if _, ok := err.(errors.HttpError); ok {
		err = nil
	}

	if err != nil {
		return
	}

	headerBytes, _ := httputil.DumpResponse(res, false)
	resHeaders = string(headerBytes)

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = errors.NewWithError("Error reading response", err)
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
