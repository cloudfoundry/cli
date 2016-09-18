package cloudcontrollerv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (client *CloudControllerClient) API() string {
	return client.cloudControllerURL
}

func (client *CloudControllerClient) APIVersion() string {
	return client.cloudControllerAPIVersion
}

func (client *CloudControllerClient) AuthorizationEndpoint() string {
	return client.authorizationEndpoint
}

func (client *CloudControllerClient) DopplerEndpoint() string {
	return client.dopplerEndpoint
}

func (client *CloudControllerClient) LoggregatorEndpoint() string {
	return client.loggregatorEndpoint
}

func (client *CloudControllerClient) RoutingEndpoint() string {
	return client.routingEndpoint
}

func (client *CloudControllerClient) TokenEndpoint() string {
	return client.tokenEndpoint
}

func (client *CloudControllerClient) Info() (APIInformation, Warnings, error) {
	response, err := http.Get(fmt.Sprintf("http://%s/v2/info", client.cloudControllerURL))
	if err != nil {
		return APIInformation{}, nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return APIInformation{}, nil, err
	}

	var info APIInformation
	err = json.Unmarshal(body, &info)
	if err != nil {
		return APIInformation{}, nil, err
	}

	return info, nil, nil
}
