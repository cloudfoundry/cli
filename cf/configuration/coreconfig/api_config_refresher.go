package coreconfig

import (
	"fmt"
	"regexp"
	"strings"
)

//go:generate counterfeiter . EndpointRepository

type EndpointRepository interface {
	GetCCInfo(string) (*CCInfo, string, error)
}

type APIConfigRefresher struct {
	EndpointRepo EndpointRepository
	Config       ReadWriter
	Endpoint     string
}

func (a APIConfigRefresher) Refresh() error {
	ccInfo, endpoint, err := a.EndpointRepo.GetCCInfo(a.Endpoint)
	if err != nil {
		return err
	}

	if endpoint != a.Config.ApiEndpoint() {
		a.Config.ClearSession()
	}

	a.Config.SetApiEndpoint(endpoint)
	a.Config.SetApiVersion(ccInfo.ApiVersion)
	a.Config.SetAuthenticationEndpoint(ccInfo.AuthorizationEndpoint)
	a.Config.SetSSHOAuthClient(ccInfo.SSHOAuthClient)
	a.Config.SetMinCliVersion(ccInfo.MinCliVersion)
	a.Config.SetMinRecommendedCliVersion(ccInfo.MinRecommendedCliVersion)
	a.Config.SetLoggregatorEndpoint(a.LoggregatorEndpoint(ccInfo, endpoint))

	//* 3/5/15: loggregator endpoint will be renamed to doppler eventually,
	//          we just have to use the loggregator endpoint as doppler for now
	a.Config.SetDopplerEndpoint(strings.Replace(a.Config.LoggregatorEndpoint(), "loggregator", "doppler", 1))
	a.Config.SetRoutingApiEndpoint(ccInfo.RoutingApiEndpoint)

	return nil
}

func (a APIConfigRefresher) LoggregatorEndpoint(ccInfo *CCInfo, endpoint string) string {
	if ccInfo.LoggregatorEndpoint == "" {
		var endpointDomainRegex = regexp.MustCompile(`^http(s?)://[^\.]+\.([^:]+)`)

		matches := endpointDomainRegex.FindStringSubmatch(endpoint)
		url := fmt.Sprintf("ws%s://loggregator.%s", matches[1], matches[2])
		if url[0:3] == "wss" {
			return url + ":443"
		}
		return url + ":80"
	}
	return ccInfo.LoggregatorEndpoint
}
