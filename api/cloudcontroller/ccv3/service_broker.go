package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// ServiceBroker represents Service Broker data
type ServiceBroker struct {
	// GUID is a unique service broker identifier.
	GUID string
	// Name is the name of the service broker.
	Name string
	// URL is the url of the service broker.
	URL string
}

// serviceBrokerPresentation represents a Cloud Controller V3 Service Broker.
type serviceBrokerPresentation struct {
	// GUID is a unique service broker identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the service broker.
	Name string `json:"name"`
	// URL is the url of the service broker.
	URL string `json:"url"`
	// Credentials contains the credentials for authenticating with the service broker.
	Credentials serviceBrokerCredentials `json:"credentials"`
	// This is the relationship for the space GUID
	Relationships *serviceBrokerRelationships `json:"relationships,omitempty"`
}

// serviceBrokerCredentials represents a data structure for the Credentials
// of V3 Service Broker.
type serviceBrokerCredentials struct {
	// Type is the type of credentials for the service broker, e.g. "basic"
	Type constant.ServiceBrokerCredentialsType `json:"type"`
	// Data is the credentials data of the service broker of a particular type.
	Data serviceBrokerCredentialsData `json:"data"`
}

// serviceBrokerCredentialsData represents a data structure for the Credentials Data
// of V3 Service Broker Credentials.
type serviceBrokerCredentialsData struct {
	// Username is the Basic Auth username for the service broker.
	Username string `json:"username"`
	// Password is the Basic Auth password for the service broker.
	Password string `json:"password"`
}

// serviceBrokerRelationships represents a data structure for the relationships data
// of V3 Service Broker Relationships.
type serviceBrokerRelationships struct {
	// Space represents the space that a space-scoped broker is in
	Space serviceBrokerRelationshipsSpace `json:"space"`
}

// serviceBrokerRelationshipsSpace represents a data structure for the relationships space data
// of V3 Service Broker Relationships.
type serviceBrokerRelationshipsSpace struct {
	// Data holds the space GUID object
	Data serviceBrokerRelationshipsSpaceData `json:"data"`
}

// serviceBrokerRelationshipsSpaceData represents a data structure for the relationships space GUID data
// of V3 Service Broker Relationships.
type serviceBrokerRelationshipsSpaceData struct {
	// GUID is the space guid associated with a space-scoped broker
	GUID string `json:"guid"`
}

// CreateServiceBroker registers a new service broker.
func (client *Client) CreateServiceBroker(name, username, password, brokerURL, spaceGUID string) (JobURL, Warnings, error) {
	bodyBytes, err := json.Marshal(newServiceBroker(name, username, password, brokerURL, spaceGUID))
	if err != nil {
		return "", nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostServiceBrokerRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)
	jobURL := response.HTTPResponse.Header.Get("Location")

	return JobURL(jobURL), response.Warnings, err
}

// GetServiceBrokers lists service brokers.
func (client *Client) GetServiceBrokers() ([]ServiceBroker, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceBrokersRequest,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullList []ServiceBroker
	warnings, err := client.paginate(request, serviceBrokerPresentation{}, func(item interface{}) error {
		if serviceBroker, ok := item.(serviceBrokerPresentation); ok {
			fullList = append(fullList, extractServiceBrokerData(serviceBroker))
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   serviceBrokerPresentation{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullList, warnings, err
}

func newServiceBroker(name, username, password, brokerURL, spaceGUID string) serviceBrokerPresentation {
	sbp := serviceBrokerPresentation{
		Name: name,
		URL:  brokerURL,
		Credentials: serviceBrokerCredentials{
			Type: constant.BasicCredentials,
			Data: serviceBrokerCredentialsData{
				Username: username,
				Password: password,
			},
		},
	}

	if spaceGUID != "" {
		sbp.Relationships = &serviceBrokerRelationships{
			Space: serviceBrokerRelationshipsSpace{
				Data: serviceBrokerRelationshipsSpaceData{
					GUID: spaceGUID,
				},
			},
		}
	}

	return sbp
}

func extractServiceBrokerData(sbp serviceBrokerPresentation) ServiceBroker {
	return ServiceBroker{
		Name: sbp.Name,
		URL:  sbp.URL,
		GUID: sbp.GUID,
	}
}
