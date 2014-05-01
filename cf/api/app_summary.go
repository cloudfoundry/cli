package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"strings"
)

type ApplicationSummaries struct {
	Apps []ApplicationFromSummary
}

func (resource ApplicationSummaries) ToModels() (apps []models.ApplicationFields) {
	for _, application := range resource.Apps {
		apps = append(apps, application.ToFields())
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

func (resource ApplicationFromSummary) ToFields() (app models.ApplicationFields) {
	app = models.ApplicationFields{}
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

func (resource ApplicationFromSummary) ToModel() (app models.Application) {
	app.ApplicationFields = resource.ToFields()
	routes := []models.RouteSummary{}
	for _, route := range resource.Routes {
		routes = append(routes, route.ToModel())
	}
	app.Routes = routes

	return
}

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainSummary
}

func (resource RouteSummary) ToModel() (route models.RouteSummary) {
	domain := models.DomainFields{}
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
	GetSummariesInCurrentSpace() (apps []models.Application, apiErr error)
	GetSummary(appGuid string) (summary models.Application, apiErr error)
}

type CloudControllerAppSummaryRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerAppSummaryRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerAppSummaryRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummariesInCurrentSpace() (apps []models.Application, apiErr error) {
	resources := new(ApplicationSummaries)

	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.ApiEndpoint(), repo.config.SpaceFields().Guid)
	apiErr = repo.gateway.GetResource(path, resources)
	if apiErr != nil {
		return
	}

	for _, resource := range resources.Apps {
		apps = append(apps, resource.ToModel())
	}
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummary(appGuid string) (summary models.Application, apiErr error) {
	path := fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.ApiEndpoint(), appGuid)
	summaryResponse := new(ApplicationFromSummary)
	apiErr = repo.gateway.GetResource(path, summaryResponse)
	if apiErr != nil {
		return
	}

	summary = summaryResponse.ToModel()
	return
}
