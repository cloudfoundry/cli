package resources

import (
	"code.cloudfoundry.org/cli/v9/types"
	"code.cloudfoundry.org/jsonry"
)

type ServiceOffering struct {
	// GUID is a unique service offering identifier.
	GUID string `json:"guid"`
	// Name is the name of the service offering.
	Name string `json:"name"`
	// Description of the service offering
	Description string `json:"description"`
	// DocumentationURL of the service offering
	DocumentationURL string `json:"documentation_url"`
	// Tags are used by apps to identify service instances.
	Tags types.OptionalStringSlice `jsonry:"tags"`
	// ServiceBrokerGUID is the guid of the service broker
	ServiceBrokerGUID string `jsonry:"relationships.service_broker.data.guid"`
	// ServiceBrokerName is the name of the service broker
	ServiceBrokerName string `json:"-"`
	// Shareable if the offering support service instance sharing
	AllowsInstanceSharing bool `json:"shareable"`

	Metadata *Metadata `json:"metadata"`
}

func (s *ServiceOffering) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
