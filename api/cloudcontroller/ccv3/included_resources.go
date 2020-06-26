package ccv3

import "code.cloudfoundry.org/cli/resources"

type IncludedResources struct {
	Users            []resources.User            `json:"users,omitempty"`
	Organizations    []resources.Organization    `json:"organizations,omitempty"`
	Spaces           []resources.Space           `json:"spaces,omitempty"`
	ServiceOfferings []resources.ServiceOffering `json:"service_offerings,omitempty"`
	ServiceBrokers   []resources.ServiceBroker   `json:"service_brokers,omitempty"`
}
