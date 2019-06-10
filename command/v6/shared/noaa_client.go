package shared

import (
	"code.cloudfoundry.org/cli/api/uaa/noaabridge"
	"crypto/tls"
	"github.com/cloudfoundry/noaa/consumer"
	"net"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/command"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
)

type RequestLoggerOutput interface {
	Start() error
	Stop() error
	DisplayType(name string, requestDate time.Time) error
	DisplayDump(dump string) error
}

type DebugPrinter struct {
	outputs []RequestLoggerOutput
}

func (p *DebugPrinter) addOutput(output RequestLoggerOutput) {
	p.outputs = append(p.outputs, output)
}

func (p DebugPrinter) Print(title string, dump string) {
	for _, output := range p.outputs {
		_ = output.Start()
		defer output.Stop()

		output.DisplayType(title, time.Now())
		output.DisplayDump(dump)
	}

}

type tokenHTTPClient struct {
	c           logcache.HTTPClient
	accessToken func() string
}

func (c *tokenHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", c.accessToken())

	return c.c.Do(req)
}

// NewLogCacheClient returns back a configured Log Cache Client.
func NewLogCacheClient(ccClient *ccv2.Client, config command.Config) *logcache.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipSSLValidation(),
		},
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   config.DialTimeout(),
		}).DialContext,
	}
	return logcache.NewClient(ccClient.LogCacheEndpoint(), logcache.WithHTTPClient(&tokenHTTPClient{
		c:           &http.Client{Transport: tr},
		accessToken: config.AccessToken,
	}))
}

// NewNOAAClient returns back a configured NOAA Client.
func NewNOAAClient(apiURL string, config command.Config, uaaClient *uaa.Client, ui command.UI) *consumer.Consumer {
	client := consumer.New(
		apiURL,
		&tls.Config{
			InsecureSkipVerify: config.SkipSSLValidation(),
		},
		http.ProxyFromEnvironment,
	)
	client.RefreshTokenFrom(noaabridge.NewTokenRefresher(uaaClient, config))
	client.SetMaxRetryCount(config.NOAARequestRetryCount())

	noaaDebugPrinter := DebugPrinter{}

	// if verbose, set debug printer on noaa client
	verbose, location := config.Verbose()

	client.SetDebugPrinter(&noaaDebugPrinter)

	if verbose {
		noaaDebugPrinter.addOutput(ui.RequestLoggerTerminalDisplay())
	}
	if location != nil {
		noaaDebugPrinter.addOutput(ui.RequestLoggerFileWriter(location))
	}

	return client
}
