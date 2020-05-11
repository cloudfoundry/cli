package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"github.com/tedsuo/rata"
	"strings"
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
	client.router = rata.NewRequestGenerator(settings.URL, internal.APIRoutes)

	client.connection = cloudcontroller.NewConnection(cloudcontroller.Config{
		DialTimeout:       settings.DialTimeout,
		SkipSSLValidation: settings.SkipSSLValidation,
	})

	for _, wrapper := range client.wrappers {
		client.connection = wrapper.Wrap(client.connection)
	}

	info, warnings, err := client.Info()
	if err != nil {
		return warnings, err
	}

	rootInfo, _, err := client.RootResponse()
	if err != nil {
		return warnings, err
	}

	client.authorizationEndpoint = info.AuthorizationEndpoint
	client.cloudControllerAPIVersion = info.APIVersion
	client.dopplerEndpoint = info.DopplerEndpoint
	//TODO Remove this condition when earliest supportest CAPI is 1.87.0
	//We have to do this because the current legacy supported CAPI version as of 2020 does not display the log cache url, this will break if a foundation on legacy CAPI have non-standard logcache urls
	if rootInfo.Links.LogCache.HREF != "" {
		client.logCacheEndpoint = rootInfo.Links.LogCache.HREF
	} else {
		client.logCacheEndpoint = strings.Replace(settings.URL, "api", "log-cache", 1)
	}
	client.minCLIVersion = info.MinCLIVersion
	client.routingEndpoint = info.RoutingEndpoint

	return warnings, nil
}
