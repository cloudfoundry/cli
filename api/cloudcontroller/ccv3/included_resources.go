package ccv3

import "code.cloudfoundry.org/cli/resources"

type IncludedResources struct {
	Apps             []resources.Application     `json:"apps"`
	Organizations    []resources.Organization    `json:"organizations"`
	Routes           []resources.Route           `json:"routes"`
	ServiceBrokers   []resources.ServiceBroker   `json:"service_brokers"`
	ServiceInstances []resources.ServiceInstance `json:"service_instances"`
	ServiceOfferings []resources.ServiceOffering `json:"service_offerings"`
	ServicePlans     []resources.ServicePlan     `json:"service_plans"`
	Spaces           []resources.Space           `json:"spaces"`
	Users            []resources.User            `json:"users"`
}
