package ccv3

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"

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

	Metadata *Metadata
}

func (s *ServiceBroker) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}

type ServiceBrokerModel struct {
	// Name is the name of the service broker.
	Name string `jsonry:"name,omitempty"`
	// URL is the url of the service broker.
	URL string `jsonry:"url,omitempty"`
	// CredentialsType is always "basic"
	CredentialsType constant.ServiceBrokerCredentialsType `jsonry:"authentication.type,omitempty"`
	// Username is the Basic Auth username for the service broker.
	Username string `jsonry:"authentication.credentials.username,omitempty"`
	// Password is the Basic Auth password for the service broker.
	Password string `jsonry:"authentication.credentials.password,omitempty"`
	// Space GUID for the space that the broker is in. Empty when not a space-scoped service broker.
	SpaceGUID string `jsonry:"relationships.space.data.guid,omitempty"`
}

func (s ServiceBrokerModel) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}

func (s ServiceBrokerModel) check() (ServiceBrokerModel, error) {
	if (s.Username == "" && s.Password != "") || (s.Username != "" && s.Password == "") {
		return s, errors.New("Incorrect usage: both username and password must be defined in order to do an update")
	}

	if s.Username != "" && s.Password != "" {
		s.CredentialsType = constant.BasicCredentials
	}

	return s, nil
}

// CreateServiceBroker registers a new service broker.
func (client *Client) CreateServiceBroker(serviceBroker ServiceBrokerModel) (JobURL, Warnings, error) {
	serviceBroker.CredentialsType = constant.BasicCredentials

	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PostServiceBrokerRequest,
		RequestBody: serviceBroker,
	})

	return jobURL, warnings, err
}

// DeleteServiceBroker deletes a named service broker
func (client *Client) DeleteServiceBroker(serviceBrokerGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceBrokerRequest,
		URIParams:   internal.Params{"service_broker_guid": serviceBrokerGUID},
	})

	return jobURL, warnings, err
}

// GetServiceBrokers lists service brokers.
func (client *Client) GetServiceBrokers(query ...Query) ([]ServiceBroker, Warnings, error) {
	var resources []ServiceBroker

	_, warnings, err := client.MakeListRequest(RequestParams{
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
	brokerUpdateRequest, err := serviceBroker.check()
	if err != nil {
		return "", nil, err
	}

	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PatchServiceBrokerRequest,
		URIParams:   internal.Params{"service_broker_guid": serviceBrokerGUID},
		RequestBody: brokerUpdateRequest,
	})

	return jobURL, warnings, err
}
