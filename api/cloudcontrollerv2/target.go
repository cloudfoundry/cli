package cloudcontrollerv2

type APIInformation struct {
	APIVersion                   string `json:"api_version"`
	AuthorizationEndpoint        string `json:"authorization_endpoint"`
	DopplerEndpoint              string `json:"doppler_logging_endpoint"`
	LoggregatorEndpoint          string `json:"logging_endpoint"`
	MinimumCLIVersion            string `json:"min_cli_version"`
	MinimumRecommendedCLIVersion string `json:"min_recommended_cli_version"`
	Name                         string `json:"name"`
	RoutingEndpoint              string `json:"routing_endpoint"`
	TokenEndpoint                string `json:"token_endpoint"`
}

func (client *CloudControllerClient) TargetCF(APIURL string, skipSSLValidation bool) (Warnings, error) {
	client.cloudControllerURL = APIURL

	client.connection = NewConnection(client.cloudControllerURL, skipSSLValidation)
	request := Request{
		RequestName: InfoRequest,
	}

	var info APIInformation
	response := Response{
		Result: &info,
	}
	err := client.connection.Make(request, &response)
	if err != nil {
		return Warnings(response.Warnings), err
	}

	client.authorizationEndpoint = info.AuthorizationEndpoint
	client.cloudControllerAPIVersion = info.APIVersion
	client.dopplerEndpoint = info.DopplerEndpoint
	client.loggregatorEndpoint = info.LoggregatorEndpoint
	client.routingEndpoint = info.RoutingEndpoint
	client.tokenEndpoint = info.TokenEndpoint

	return Warnings(response.Warnings), nil
}
