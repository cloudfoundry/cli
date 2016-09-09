package cloudcontrollerv2

type APIInformation struct {
	Name                         string `json:"name"`
	AuthorizationEndpoint        string `json:"authorization_endpoint"`
	TokenEndpoint                string `json:"token_endpoint"`
	MinimumCLIVersion            string `json:"min_cli_version"`
	MinimumRecommendedCLIVersion string `json:"min_recommended_cli_version"`
	APIVersion                   string `json:"api_version"`
	LoggregatorEndpoint          string `json:"logging_endpoint"`
	DopplerEndpoint              string `json:"doppler_logging_endpoint"`
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

	client.cloudControllerAPIVersion = info.APIVersion
	client.authorizationEndpoint = info.AuthorizationEndpoint
	client.loggregatorEndpoint = info.LoggregatorEndpoint
	client.dopplerEndpoint = info.DopplerEndpoint
	client.tokenEndpoint = info.TokenEndpoint

	return Warnings(response.Warnings), nil
}
