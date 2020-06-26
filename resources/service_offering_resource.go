package resources

import (
	"code.cloudfoundry.org/jsonry"
)

type ServiceOffering struct {
	// GUID is a unique service offering identifier.
	GUID string `json:"guid"`
	// Name is the name of the service offering.
	Name string `json:"name"`
	// Description of the service offering
	Description string `json:"description"`
	// ServiceBrokerGUID is the guid of the service broker
	ServiceBrokerGUID string `jsonry:"relationships.service_broker.data.guid"`
	// ServiceBrokerName is the name of the service broker
	ServiceBrokerName string `json:"-"`

	Metadata *Metadata `json:"metadata"`
}

func (s *ServiceOffering) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
