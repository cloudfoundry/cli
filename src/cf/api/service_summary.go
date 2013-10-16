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

type ServiceInstanceSummaryApp struct {
	Name         string
	ServiceNames []string `json:"service_names"`
}

type ServiceInstanceSummary struct {
	Name        string
	ServicePlan ServicePlanSummary `json:"service_plan"`
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
	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.Target, repo.config.Space.Guid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	response := new(ServiceInstancesSummaries)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)

	if apiResponse.IsNotSuccessful() {
		return
	}

	instances = extractServiceInstancesFromSummary(response.ServiceInstances, response.Apps)

	return
}

func extractServiceInstancesFromSummary(instanceSummaries []ServiceInstanceSummary, apps []ServiceInstanceSummaryApp) (instances []cf.ServiceInstance) {
	for _, instanceSummary := range instanceSummaries {
		applicationNames := findApplicationNamesForInstance(instanceSummary.Name, apps)

		planSummary := instanceSummary.ServicePlan
		offeringSummary := planSummary.ServiceOffering

		serviceOffering := cf.ServiceOffering{
			Label:    offeringSummary.Label,
			Provider: offeringSummary.Provider,
			Version:  offeringSummary.Version,
		}

		servicePlan := cf.ServicePlan{
			Name:            planSummary.Name,
			Guid:            planSummary.Guid,
			ServiceOffering: serviceOffering,
		}

		instance := cf.ServiceInstance{
			Name:             instanceSummary.Name,
			ServicePlan:      servicePlan,
			ApplicationNames: applicationNames,
		}

		instances = append(instances, instance)
	}

	return
}

func findApplicationNamesForInstance(instanceName string, apps []ServiceInstanceSummaryApp) (applicationNames []string) {
	for _, app := range apps {
		for _, name := range app.ServiceNames {
			if name == instanceName {
				applicationNames = append(applicationNames, app.Name)
			}
		}
	}

	return
}
