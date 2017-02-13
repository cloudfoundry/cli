package shared

import (
	"crypto/tls"
	"net/http"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/noaabridge"
	"code.cloudfoundry.org/cli/command"
	"github.com/cloudfoundry/noaa/consumer"
)

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

	return client
}
