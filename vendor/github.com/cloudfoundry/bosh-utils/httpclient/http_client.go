package httpclient

import (
	"net/http"
	"strings"

	"errors"
	"regexp"

	"net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type HTTPClient struct {
	client Client
	logger boshlog.Logger
	logTag string
	opts   Opts
}

type Opts struct {
	NoRedactUrlQuery bool
}

func NewHTTPClient(client Client, logger boshlog.Logger) *HTTPClient {
	return &HTTPClient{
		client: client,
		logger: logger,
		logTag: "httpClient",
	}
}

func NewHTTPClientOpts(client Client, logger boshlog.Logger, opts Opts) *HTTPClient {
	return &HTTPClient{
		client: client,
		logger: logger,
		logTag: "httpClient",
		opts:   opts,
	}
}

func (c *HTTPClient) Post(endpoint string, payload []byte) (*http.Response, error) {
	return c.PostCustomized(endpoint, payload, nil)
}

func (c *HTTPClient) PostCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error) {
	postPayload := strings.NewReader(string(payload))

	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	c.logger.Debug(c.logTag, "Sending POST request to endpoint '%s'", redactedEndpoint)

	request, err := http.NewRequest("POST", endpoint, postPayload)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating POST request")
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(scrubErrorOutput(err), "Performing POST request")
	}

	return response, nil
}

func (c *HTTPClient) Put(endpoint string, payload []byte) (*http.Response, error) {
	return c.PutCustomized(endpoint, payload, nil)
}

func (c *HTTPClient) PutCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error) {
	putPayload := strings.NewReader(string(payload))

	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	c.logger.Debug(c.logTag, "Sending PUT request to endpoint '%s'", redactedEndpoint)

	request, err := http.NewRequest("PUT", endpoint, putPayload)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating PUT request")
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(scrubErrorOutput(err), "Performing PUT request")
	}

	return response, nil
}

func (c *HTTPClient) Get(endpoint string) (*http.Response, error) {
	return c.GetCustomized(endpoint, nil)
}

func (c *HTTPClient) GetCustomized(endpoint string, f func(*http.Request)) (*http.Response, error) {
	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	c.logger.Debug(c.logTag, "Sending GET request to endpoint '%s'", redactedEndpoint)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating GET request")
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(scrubErrorOutput(err), "Performing GET request")
	}

	return response, nil
}

func (c *HTTPClient) Delete(endpoint string) (*http.Response, error) {
	return c.DeleteCustomized(endpoint, nil)
}

func (c *HTTPClient) DeleteCustomized(endpoint string, f func(*http.Request)) (*http.Response, error) {
	redactedEndpoint := endpoint

	if !c.opts.NoRedactUrlQuery {
		redactedEndpoint = scrubEndpointQuery(endpoint)
	}

	c.logger.Debug(c.logTag, "Sending DELETE request with endpoint %s", redactedEndpoint)

	request, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating DELETE request")
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing DELETE request")
	}
	return response, nil
}

var scrubUserinfoRegex = regexp.MustCompile("(https?://.*:).*@")

func scrubEndpointQuery(endpoint string) string {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "error occurred parsing endpoing"
	}

	query := parsedURL.Query()
	for key, _ := range query {
		query[key] = []string{"<redacted>"}
	}

	parsedURL.RawQuery = query.Encode()

	unescapedEndpoint, _ := url.QueryUnescape(parsedURL.String())
	return unescapedEndpoint
}

func scrubErrorOutput(err error) error {
	errorMsg := err.Error()
	errorMsg = scrubUserinfoRegex.ReplaceAllString(errorMsg, "$1<redacted>@")

	return errors.New(errorMsg)
}
