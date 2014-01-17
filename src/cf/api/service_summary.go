package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

type ServiceInstancesSummaries struct {
	Apps             []ServiceInstanceSummaryApp
	ServiceInstances []ServiceInstanceSummary `json:"services"`
}

func (resource ServiceInstancesSummaries) ToModels() (instances []cf.ServiceInstance) {
	for _, instanceSummary := range resource.ServiceInstances {
		applicationNames := resource.findApplicationNamesForInstance(instanceSummary.Name)

		planSummary := instanceSummary.ServicePlan
		servicePlan := cf.ServicePlanFields{}
		servicePlan.Name = planSummary.Name
		servicePlan.Guid = planSummary.Guid

		offeringSummary := planSummary.ServiceOffering
		serviceOffering := cf.ServiceOfferingFields{}
		serviceOffering.Label = offeringSummary.Label
		serviceOffering.Provider = offeringSummary.Provider
		serviceOffering.Version = offeringSummary.Version

		instance := cf.ServiceInstance{}
		instance.Name = instanceSummary.Name
		instance.Guid = instanceSummary.Guid
		instance.DashboardURL = instanceSummary.DashboardURL
		instance.ApplicationNames = applicationNames
		instance.ServicePlan = servicePlan
		instance.ServiceOffering = serviceOffering

		instances = append(instances, instance)
	}

	return
}

func (resource ServiceInstancesSummaries) findApplicationNamesForInstance(instanceName string) (applicationNames []string) {
	for _, app := range resource.Apps {
		for _, name := range app.ServiceNames {
			if name == instanceName {
				applicationNames = append(applicationNames, app.Name)
			}
		}
	}

	return
}

type ServiceInstanceSummaryApp struct {
	Name         string
	ServiceNames []string `json:"service_names"`
}

type ServiceInstanceSummary struct {
	Name         string
	Guid         string
	DashboardURL string             `json:"dashboard_url"`
	ServicePlan  ServicePlanSummary `json:"service_plan"`
}

type ServicePlanSummary struct {
	Name            string
	Guid            string
	ServiceOffering ServiceOfferingSummary `json:"service"`
}

type ServiceOfferingSummary struct {
	Label    string
	Provider string
	Version  string
}

type ServiceSummaryRepository interface {
	GetSummariesInCurrentSpace() (instances []cf.ServiceInstance, apiResponse net.ApiResponse)
}

type CloudControllerServiceSummaryRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerServiceSummaryRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerServiceSummaryRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceSummaryRepository) GetSummariesInCurrentSpace() (instances []cf.ServiceInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.Target, repo.config.SpaceFields.Guid)
	resource := new(ServiceInstancesSummaries)

	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instances = resource.ToModels()

	return
}
