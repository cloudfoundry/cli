package ccv3

import (
	"time"
)

// TargetSettings represents configuration for establishing a connection to the
// Cloud Controller server.
type TargetSettings struct {
	// DialTimeout is the DNS timeout used to make all requests to the Cloud
	// Controller.
	DialTimeout time.Duration

	// SkipSSLValidation controls whether a client verifies the server's
	// certificate chain and host name. If SkipSSLValidation is true, TLS accepts
	// any certificate presented by the server and any host name in that
	// certificate for *all* client requests going forward.
	//
	// In this mode, TLS is susceptible to man-in-the-middle attacks. This should
	// be used only for testing.
	SkipSSLValidation bool

	// URL is a fully qualified URL to the Cloud Controller API.
	URL string
}

// TargetCF sets the client to use the Cloud Controller specified in the
// configuration. Any other configuration is also applied to the client.
func (client *Client) TargetCF(settings TargetSettings) (Warnings, error) {
	client.cloudControllerURL = settings.URL
	client.InitializeConnection(settings)

	apiInfo, resourceLinks, warnings, err := client.GetInfo()
	if err != nil {
		return warnings, err
	}

	client.Info = apiInfo

	resources := map[string]string{}
	for resource, link := range resourceLinks {
		resources[resource] = link.HREF
	}
	client.InitializeRouter(resources)

	return warnings, nil
}
