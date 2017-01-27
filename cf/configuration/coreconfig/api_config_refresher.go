package coreconfig

import (
	"strings"

	. "code.cloudfoundry.org/cli/cf/i18n"
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

func (a APIConfigRefresher) Refresh() (Warning, error) {
	ccInfo, endpoint, err := a.EndpointRepo.GetCCInfo(a.Endpoint)
	if err != nil {
		return nil, err
	}

	if endpoint != a.Config.APIEndpoint() {
		a.Config.ClearSession()
	}

	a.Config.SetAPIEndpoint(endpoint)
	a.Config.SetAPIVersion(ccInfo.APIVersion)
	a.Config.SetAuthenticationEndpoint(ccInfo.AuthorizationEndpoint)
	a.Config.SetSSHOAuthClient(ccInfo.SSHOAuthClient)
	a.Config.SetMinCLIVersion(ccInfo.MinCLIVersion)
	a.Config.SetMinRecommendedCLIVersion(ccInfo.MinRecommendedCLIVersion)

	a.Config.SetDopplerEndpoint(ccInfo.DopplerEndpoint)
	a.Config.SetRoutingAPIEndpoint(ccInfo.RoutingAPIEndpoint)

	if !strings.HasPrefix(endpoint, "https://") {
		return new(insecureWarning), nil
	}
	return nil, nil
}

type Warning interface {
	Warn() string
}

type insecureWarning struct{}

func (w insecureWarning) Warn() string {
	return T("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended\n")
}
