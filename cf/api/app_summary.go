package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
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
	Guid               string
	Name               string
	Routes             []RouteSummary
	Services           []ServicePlanSummary
	Diego              bool `json:"diego,omitempty"`
	RunningInstances   int  `json:"running_instances"`
	Memory             int64
	Instances          int
	DiskQuota          int64 `json:"disk_quota"`
	Urls               []string
	EnvironmentVars    map[string]interface{} `json:"environment_json,omitempty"`
	HealthCheckTimeout int                    `json:"health_check_timeout"`
	State              string
	SpaceGuid          string     `json:"space_guid"`
	PackageUpdatedAt   *time.Time `json:"package_updated_at"`
}

func (resource ApplicationFromSummary) ToFields() (app models.ApplicationFields) {
	app = models.ApplicationFields{}
	app.Guid = resource.Guid
	app.Name = resource.Name
	app.Diego = resource.Diego
	app.State = strings.ToLower(resource.State)
	app.InstanceCount = resource.Instances
	app.DiskQuota = resource.DiskQuota
	app.RunningInstances = resource.RunningInstances
	app.Memory = resource.Memory
	app.SpaceGuid = resource.SpaceGuid
	app.PackageUpdatedAt = resource.PackageUpdatedAt
	app.HealthCheckTimeout = resource.HealthCheckTimeout

	return
}

func (resource ApplicationFromSummary) ToModel() (app models.Application) {
	app.ApplicationFields = resource.ToFields()

	routes := []models.RouteSummary{}
	for _, route := range resource.Routes {
		routes = append(routes, route.ToModel())
	}
	app.Routes = routes

	services := []models.ServicePlanSummary{}
	for _, service := range resource.Services {
		services = append(services, service.ToModel())
	}

	app.EnvironmentVars = resource.EnvironmentVars
	app.Routes = routes
	app.Services = services

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

func (resource ServicePlanSummary) ToModel() (route models.ServicePlanSummary) {
	route.Guid = resource.Guid
	route.Name = resource.Name
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
	GetSpaceSummaries(spaceGuid string) (apps []models.Application, apiErr error)
}

type CloudControllerAppSummaryRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerAppSummaryRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerAppSummaryRepository) {
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

func (repo CloudControllerAppSummaryRepository) GetSpaceSummaries(spaceGuid string) (apps []models.Application, apiErr error) {
	resources := new(ApplicationSummaries)
	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.ApiEndpoint(), spaceGuid)
	apiErr = repo.gateway.GetResource(path, resources)
	if apiErr != nil {
		return
	}
	for _, resource := range resources.Apps {
		apps = append(apps, resource.ToModel())
	}
	return
}
