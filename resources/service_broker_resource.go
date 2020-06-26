package resources

import (
	"code.cloudfoundry.org/jsonry"
)

type ServiceBrokerCredentialsType string

const (
	ServiceBrokerBasicCredentials ServiceBrokerCredentialsType = "basic"
)

type ServiceBroker struct {
	// GUID is a unique service broker identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the service broker.
	Name string `json:"name,omitempty"`
	// URL is the url of the service broker.
	URL string `json:"url,omitempty"`
	// CredentialsType is always "basic"
	CredentialsType ServiceBrokerCredentialsType `jsonry:"authentication.type,omitempty"`
	// Username is the Basic Auth username for the service broker.
	Username string `jsonry:"authentication.credentials.username,omitempty"`
	// Password is the Basic Auth password for the service broker.
	Password string `jsonry:"authentication.credentials.password,omitempty"`
	// Space GUID for the space that the broker is in. Empty when not a space-scoped service broker.
	SpaceGUID string `jsonry:"relationships.space.data.guid,omitempty"`

	Metadata *Metadata `json:"metadata,omitempty"`
}

func (s ServiceBroker) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(s)
}

func (s *ServiceBroker) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
