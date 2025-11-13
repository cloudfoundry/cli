package ccv3

import "code.cloudfoundry.org/cli/v8/resources"

type IncludedResources struct {
	Users            []resources.User            `json:"users,omitempty"`
	Organizations    []resources.Organization    `json:"organizations,omitempty"`
	Spaces           []resources.Space           `json:"spaces,omitempty"`
	ServiceInstances []resources.ServiceInstance `json:"service_instances,omitempty"`
	ServiceOfferings []resources.ServiceOffering `json:"service_offerings,omitempty"`
	ServiceBrokers   []resources.ServiceBroker   `json:"service_brokers,omitempty"`
	ServicePlans     []resources.ServicePlan     `json:"service_plans,omitempty"`
	Apps             []resources.Application     `json:"apps,omitempty"`
}

func (i *IncludedResources) Merge(resources IncludedResources) {
	i.Apps = append(i.Apps, resources.Apps...)
	i.Users = append(i.Users, resources.Users...)
	i.Organizations = append(i.Organizations, resources.Organizations...)
	i.Spaces = append(i.Spaces, resources.Spaces...)
	i.ServiceBrokers = append(i.ServiceBrokers, resources.ServiceBrokers...)
	i.ServiceInstances = append(i.ServiceInstances, resources.ServiceInstances...)
	i.ServiceOfferings = append(i.ServiceOfferings, resources.ServiceOfferings...)
	i.ServicePlans = append(i.ServicePlans, resources.ServicePlans...)
}
