package logcache

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	logcache "code.cloudfoundry.org/go-log-cache/v2"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/shared"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/util"
)

type RequestLoggerOutput interface {
	Start() error
	Stop() error

	DisplayType(name string, requestDate time.Time) error
	DisplayDump(dump string) error

	DisplayHost(name string) error
	DisplayRequestHeader(method string, uri string, httpProtocol string) error
	DisplayResponseHeader(httpProtocol string, status string) error
}

type DebugPrinter struct {
	outputs []RequestLoggerOutput
}

func (p DebugPrinter) PrintError(err error) {
	for _, output := range p.outputs {
		_ = output.Start()                          // nolint
		_ = output.DisplayType("ERROR", time.Now()) // nolint
		_ = output.DisplayDump(err.Error())         // nolint
		_ = output.Stop()                           // nolint
	}
}

func (p DebugPrinter) PrintRequest(req *http.Request) {
	for _, output := range p.outputs {
		_ = output.Start()                                                           // nolint
		_ = output.DisplayType("REQUEST", time.Now())                                // nolint
		_ = output.DisplayRequestHeader(req.Method, req.URL.RequestURI(), req.Proto) // nolint
		_ = output.DisplayHost(req.URL.Host)                                         // nolint
		_ = output.DisplayDump(headersString(req.Header))                            // nolint
		_ = output.Stop()                                                            // nolint
	}
}

func (p DebugPrinter) PrintResponse(resp *http.Response) {
	for _, output := range p.outputs {
		_ = output.Start()                                        // nolint
		_ = output.DisplayType("RESPONSE", time.Now())            // nolint
		_ = output.DisplayResponseHeader(resp.Proto, resp.Status) // nolint
		_ = output.DisplayDump(headersString(resp.Header))        // nolint
		_ = output.Stop()                                         // nolint
	}
}

func (p *DebugPrinter) addOutput(output RequestLoggerOutput) {
	p.outputs = append(p.outputs, output)
}

type userAgentHTTPClient struct {
	c         logcache.HTTPClient
	userAgent string
}

func (c *userAgentHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", c.userAgent)
	return c.c.Do(req)
}

type tokenHTTPClient struct {
	c           logcache.HTTPClient
	accessToken func() string
}

func (c *tokenHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", c.accessToken())
	return c.c.Do(req)
}

type httpDebugClient struct {
	printer DebugPrinter
	c       logcache.HTTPClient
}

func (c *httpDebugClient) Do(req *http.Request) (*http.Response, error) {
	c.printer.PrintRequest(req)

	resp, err := c.c.Do(req)
	if err != nil {
		c.printer.PrintError(err)
		return nil, err
	}

	c.printer.PrintResponse(resp)

	return resp, err
}

// NewClient returns back a configured Log Cache Client.
func NewClient(logCacheEndpoint string, config command.Config, ui command.UI, k8sConfigGetter v7action.KubernetesConfigGetter) (*logcache.Client, error) {
	var tr http.RoundTripper = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: util.NewTLSConfig(nil, config.SkipSSLValidation()),
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   config.DialTimeout(),
		}).DialContext,
	}

	if config.IsCFOnK8s() {
		var err error
		tr, err = shared.WrapForCFOnK8sAuth(config, k8sConfigGetter, tr)
		if err != nil {
			return nil, err
		}
	}

	var client logcache.HTTPClient // nolint
	client = &userAgentHTTPClient{
		c:         &http.Client{Transport: tr},
		userAgent: fmt.Sprintf("%s/%s (%s; %s %s)", config.BinaryName(), config.BinaryVersion(), runtime.Version(), runtime.GOARCH, runtime.GOOS),
	}

	verbose, location := config.Verbose()
	if verbose && ui != nil {
		printer := DebugPrinter{}
		printer.addOutput(ui.RequestLoggerTerminalDisplay())
		if location != nil {
			printer.addOutput(ui.RequestLoggerFileWriter(location))
		}

		client = &httpDebugClient{printer: printer, c: client}
	}

	if !config.IsCFOnK8s() {
		client = &tokenHTTPClient{
			c:           client,
			accessToken: config.AccessToken,
		}
	}

	return logcache.NewClient(
		logCacheEndpoint,
		logcache.WithHTTPClient(client),
	), nil
}

func headersString(header http.Header) string {
	var result string
	for name, values := range header {
		result += name + ": " + strings.Join(values, ", ") + "\n"
	}
	return result
}
