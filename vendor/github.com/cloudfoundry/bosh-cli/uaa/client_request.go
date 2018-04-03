package uaa

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type ClientRequest struct {
	endpoint     string
	client       string
	clientSecret string
	httpClient   *httpclient.HTTPClient
	logger       boshlog.Logger
}

func NewClientRequest(
	endpoint string,
	client string,
	clientSecret string,
	httpClient *httpclient.HTTPClient,
	logger boshlog.Logger,
) ClientRequest {
	return ClientRequest{
		endpoint:     endpoint,
		client:       client,
		clientSecret: clientSecret,
		httpClient:   httpClient,
		logger:       logger,
	}
}

func (r ClientRequest) Get(path string, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
		req.SetBasicAuth(r.client, r.clientSecret)
	}

	resp, err := r.httpClient.GetCustomized(url, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request GET '%s'", url)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling UAA response")
	}

	return nil
}

func (r ClientRequest) Post(path string, payload []byte, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(r.client, r.clientSecret)
	}

	resp, err := r.httpClient.PostCustomized(url, payload, setHeaders)

	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request POST '%s'", url)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling UAA response")
	}

	return nil
}

func (r ClientRequest) readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading UAA response")
	}

	if resp.StatusCode != http.StatusOK {
		msg := "UAA responded with non-successful status code '%d' response '%s'"
		return nil, bosherr.Errorf(msg, resp.StatusCode, respBody)
	}

	return respBody, nil
}
