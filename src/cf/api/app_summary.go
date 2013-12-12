package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type ApplicationSummaries struct {
	Apps []ApplicationFromSummary
}

func (resource ApplicationSummaries) ToModels() (apps []cf.ApplicationFields) {
	for _, appSummary := range resource.Apps {
		apps = append(apps, appSummary.ToFields())
	}
	return
}

type ApplicationFromSummary struct {
	Guid             string
	Name             string
	Routes           []RouteSummary
	RunningInstances int `json:"running_instances"`
	Memory           uint64
	Instances        int
	DiskQuota        uint64 `json:"disk_quota"`
	Urls             []string
	State            string
	SpaceGuid        string `json:"space_guid"`
}

func (resource ApplicationFromSummary) ToFields() (app cf.ApplicationFields) {
	app = cf.ApplicationFields{}
	app.Guid = resource.Guid
	app.Name = resource.Name
	app.State = strings.ToLower(resource.State)
	app.InstanceCount = resource.Instances
	app.DiskQuota = resource.DiskQuota
	app.RunningInstances = resource.RunningInstances
	app.Memory = resource.Memory
	app.SpaceGuid = resource.SpaceGuid

	return
}

func (resource ApplicationFromSummary) ToModel() (app cf.AppSummary) {
	app.ApplicationFields = resource.ToFields()
	routes := []cf.RouteSummary{}
	for _, route := range resource.Routes {
		routes = append(routes, route.ToModel())
	}
	app.RouteSummaries = routes

	return
}

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainSummary
}

func (resource RouteSummary) ToModel() (route cf.RouteSummary) {
	domain := cf.DomainFields{}
	domain.Guid = resource.Domain.Guid
	domain.Name = resource.Domain.Name
	domain.Shared = resource.Domain.OwningOrganizationGuid != ""

	route.Guid = resource.Guid
	route.Host = resource.Host
	route.Domain = domain
	return
}

type DomainSummary struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
}

type AppSummaryRepository interface {
	GetSummariesInCurrentSpace() (apps []cf.AppSummary, apiResponse net.ApiResponse)
	GetSummary(appGuid string) (summary cf.AppSummary, apiResponse net.ApiResponse)
}

type CloudControllerAppSummaryRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerAppSummaryRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerAppSummaryRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummariesInCurrentSpace() (apps []cf.AppSummary, apiResponse net.ApiResponse) {
	resources := new(ApplicationSummaries)

	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.Target, repo.config.SpaceFields.Guid)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, resource := range resources.Apps {
		apps = append(apps, resource.ToModel())
	}
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummary(appGuid string) (summary cf.AppSummary, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, appGuid)
	summaryResponse := new(ApplicationFromSummary)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, summaryResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	summary = summaryResponse.ToModel()
	return
}
