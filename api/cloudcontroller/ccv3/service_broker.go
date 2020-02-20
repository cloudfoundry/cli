package ccv3

import (
	"errors"

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
	// Status is the state of the service broker.
	Status string

	Metadata *Metadata
}

type ServiceBrokerModel struct {
	// Name is the name of the service broker.
	Name string
	// URL is the url of the service broker.
	URL string
	// Username is the Basic Auth username for the service broker.
	Username string
	// Password is the Basic Auth password for the service broker.
	Password string
	// Space GUID for the space that the broker is in. Empty when not a space-scoped service broker.
	SpaceGUID string
}

// serviceBrokerRequest represents a Cloud Controller V3 Service Broker (when creating and updating).
type serviceBrokerRequest struct {
	// GUID is a unique service broker identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the service broker.
	Name string `json:"name,omitempty"`
	// URL is the url of the service broker.
	URL string `json:"url,omitempty"`
	// Authentication contains the authentication for authenticating with the service broker.
	Authentication *serviceBrokerAuthentication `json:"authentication,omitempty"`
	// This is the relationship for the space GUID
	Relationships *serviceBrokerRelationships `json:"relationships,omitempty"`
}

// serviceBrokerResponse represents a Cloud Controller V3 Service Broker (when reading).
type serviceBrokerResponse struct {
	// GUID is a unique service broker identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the service broker.
	Name string `json:"name"`
	// URL is the url of the service broker.
	URL string `json:"url"`
	// Status is the state of the service broker.
	Status string `json:"status,omitempty"`
	// Authentication contains the authentication for authenticating with the service broker.
	Authentication serviceBrokerAuthentication `json:"authentication"`
	// Metadata is used for custom tagging of API resources
	Metadata *Metadata
	// This is the relationship for the space GUID
	Relationships *serviceBrokerRelationships `json:"relationships,omitempty"`
}

// serviceBrokerAuthentication represents a data structure for the Credentials
// of V3 Service Broker.
type serviceBrokerAuthentication struct {
	// Type is the type of authentication for the service broker, e.g. "basic"
	Type constant.ServiceBrokerCredentialsType `json:"type"`
	// Data is the authentication data of the service broker of a particular type.
	Credentials serviceBrokerBasicAuthCredentials `json:"credentials"`
}

// serviceBrokerBasicAuthCredentials represents a data structure for the Credentials Data
// of V3 Service Broker Credentials.
type serviceBrokerBasicAuthCredentials struct {
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
func (client *Client) CreateServiceBroker(serviceBroker ServiceBrokerModel) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.PostServiceBrokerRequest,
		RequestBody: newServiceBroker(serviceBroker),
	})

	return jobURL, warnings, err
}

// DeleteServiceBroker deletes a named service broker
func (client *Client) DeleteServiceBroker(serviceBrokerGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteServiceBrokerRequest,
		URIParams:   internal.Params{"service_broker_guid": serviceBrokerGUID},
	})

	return jobURL, warnings, err
}

// GetServiceBrokers lists service brokers.
func (client *Client) GetServiceBrokers(query ...Query) ([]ServiceBroker, Warnings, error) {
	var resources []ServiceBroker

	_, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetServiceBrokersRequest,
		Query:        query,
		ResponseBody: ServiceBroker{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(ServiceBroker))
			return nil
		},
	})

	return resources, warnings, err
}

// UpdateServiceBroker updates an existing service broker.
func (client *Client) UpdateServiceBroker(serviceBrokerGUID string, serviceBroker ServiceBrokerModel) (JobURL, Warnings, error) {
	brokerUpdateRequest, err := newUpdateServiceBroker(serviceBroker)
	if err != nil {
		return "", nil, err
	}

	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.PatchServiceBrokerRequest,
		URIParams:   internal.Params{"service_broker_guid": serviceBrokerGUID},
		RequestBody: brokerUpdateRequest,
	})

	return jobURL, warnings, err
}

func newServiceBroker(serviceBroker ServiceBrokerModel) serviceBrokerRequest {
	serviceBrokerRequest := serviceBrokerRequest{
		Name: serviceBroker.Name,
		URL:  serviceBroker.URL,
		Authentication: &serviceBrokerAuthentication{
			Type: constant.BasicCredentials,
			Credentials: serviceBrokerBasicAuthCredentials{
				Username: serviceBroker.Username,
				Password: serviceBroker.Password,
			},
		},
	}

	if serviceBroker.SpaceGUID != "" {
		serviceBrokerRequest.Relationships = &serviceBrokerRelationships{
			Space: serviceBrokerRelationshipsSpace{
				Data: serviceBrokerRelationshipsSpaceData{
					GUID: serviceBroker.SpaceGUID,
				},
			},
		}
	}

	return serviceBrokerRequest
}

func newUpdateServiceBroker(serviceBroker ServiceBrokerModel) (serviceBrokerRequest, error) {
	name := serviceBroker.Name
	username := serviceBroker.Username
	password := serviceBroker.Password
	brokerURL := serviceBroker.URL
	if (username == "" && password != "") || (username != "" && password == "") {
		return serviceBrokerRequest{}, errors.New("Incorrect usage: both username and password must be defined in order to do an update")
	}
	request := serviceBrokerRequest{
		Name: name,
		URL:  brokerURL,
	}

	if username != "" && password != "" {
		request.Authentication = &serviceBrokerAuthentication{
			Type: constant.BasicCredentials,
			Credentials: serviceBrokerBasicAuthCredentials{
				Username: username,
				Password: password,
			},
		}
	}

	return request, nil
}

func extractServiceBrokerData(response serviceBrokerResponse) ServiceBroker {
	return ServiceBroker{
		Name:     response.Name,
		URL:      response.URL,
		GUID:     response.GUID,
		Status:   response.Status,
		Metadata: response.Metadata,
	}
}
