package ccv3

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// TargetSettings represents configuration for establishing a connection to the
// Cloud Controller server.
type TargetSettings struct {
	DialTimeout       time.Duration
	SkipSSLValidation bool

	URL string
}

// TargetCF sets the client to use the Cloud Controller at the fully qualified
// API URL. skipSSLValidation controls whether a client verifies the server's
// certificate chain and host name. If skipSSLValidation is true, TLS accepts
// any certificate presented by the server and any host name in that
// certificate for *all* client requests going forward.
//
// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
// be used only for testing.
func (client *Client) TargetCF(settings TargetSettings) (Warnings, error) {
	client.cloudControllerURL = settings.URL

	client.connection = cloudcontroller.NewConnection(cloudcontroller.Config{
		DialTimeout:       settings.DialTimeout,
		SkipSSLValidation: settings.SkipSSLValidation,
	})
	client.WrapConnection(newErrorWrapper()) //Pretty Sneaky, Sis..

	apiInfo, _, warnings, err := client.Info()
	if err != nil {
		return warnings, err
	}

	client.APIInfo = apiInfo

	return warnings, nil
}
