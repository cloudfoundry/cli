package ccv3

import "code.cloudfoundry.org/cli/resources"

type IncludedResources struct {
	Users            []resources.User         `json:"users,omitempty"`
	Organizations    []resources.Organization `json:"organizations,omitempty"`
	Spaces           []Space                  `json:"spaces,omitempty"`
	ServiceOfferings []ServiceOffering        `json:"service_offerings,omitempty"`
	ServiceBrokers   []ServiceBroker          `json:"service_brokers,omitempty"`
}
