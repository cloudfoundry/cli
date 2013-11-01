package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strconv"
	"strings"
)

type ApplicationSummaries struct {
	Apps []ApplicationSummary
}

type ApplicationSummary struct {
	Guid             string
	Name             string
	Routes           []RouteSummary
	RunningInstances int `json:"running_instances"`
	Memory           uint64
	Instances        int
	DiskQuota        uint64 `json:"disk_quota"`
	Urls             []string
	State            string
}

type RouteSummary struct {
	Guid   string
	Host   string
	Domain DomainSummary
}

type DomainSummary struct {
	Guid string
	Name string
}

type AppSummaryRepository interface {
	GetSummariesInCurrentSpace() (apps []cf.Application, apiResponse net.ApiResponse)
	GetSummary(app cf.Application) (summary cf.AppSummary, apiResponse net.ApiResponse)
}

type CloudControllerAppSummaryRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
	appRepo ApplicationRepository
}

func NewCloudControllerAppSummaryRepository(config *configuration.Configuration, gateway net.Gateway, appRepo ApplicationRepository) (repo CloudControllerAppSummaryRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.appRepo = appRepo
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummariesInCurrentSpace() (apps []cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.Target, repo.config.Space.Guid)
	resource := new(ApplicationSummaries)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, appSummary := range resource.Apps {
		app := repo.appFromSummary(appSummary)
		apps = append(apps, app)
	}

	return
}

func (repo CloudControllerAppSummaryRepository) appFromSummary(appSummary ApplicationSummary) (app cf.Application) {
	app = cf.Application{
		Name:             appSummary.Name,
		Guid:             appSummary.Guid,
		Routes:           repo.routesFromSummary(appSummary),
		State:            strings.ToLower(appSummary.State),
		Instances:        appSummary.Instances,
		DiskQuota:        appSummary.DiskQuota,
		RunningInstances: appSummary.RunningInstances,
		Memory:           appSummary.Memory,
	}
	return
}

func (repo CloudControllerAppSummaryRepository) routesFromSummary(appSummary ApplicationSummary) (routes []cf.Route) {
	for _, routeSummary := range appSummary.Routes {
		routes = append(routes, cf.Route{
			Guid: routeSummary.Guid,
			Host: routeSummary.Host,
			Domain: cf.Domain{
				Guid: routeSummary.Domain.Guid,
				Name: routeSummary.Domain.Name,
			},
		})
	}
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummary(app cf.Application) (summary cf.AppSummary, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, app.Guid)
	summaryResponse := new(ApplicationSummary)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, summaryResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	app = repo.appFromSummary(*summaryResponse)

	summary.App = app

	instances, apiResponse := repo.appRepo.GetInstances(app)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instances, apiResponse = repo.updateInstancesWithStats(app, instances)
	if apiResponse.IsNotSuccessful() {
		return
	}

	summary.Instances = instances

	return
}

type StatsApiResponse map[string]InstanceStatsApiResponse

type InstanceStatsApiResponse struct {
	Stats struct {
		DiskQuota uint64 `json:"disk_quota"`
		MemQuota  uint64 `json:"mem_quota"`
		Usage     struct {
			Cpu  float64
			Disk uint64
			Mem  uint64
		}
	}
}

func (repo CloudControllerAppSummaryRepository) updateInstancesWithStats(app cf.Application, instances []cf.ApplicationInstance) (updatedInst []cf.ApplicationInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", repo.config.Target, app.Guid)
	statsResponse := StatsApiResponse{}
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, &statsResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedInst = make([]cf.ApplicationInstance, len(statsResponse), len(statsResponse))
	for k, v := range statsResponse {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}

		instance := instances[index]
		instance.CpuUsage = v.Stats.Usage.Cpu
		instance.DiskQuota = v.Stats.DiskQuota
		instance.DiskUsage = v.Stats.Usage.Disk
		instance.MemQuota = v.Stats.MemQuota
		instance.MemUsage = v.Stats.Usage.Mem

		updatedInst[index] = instance
	}
	return
}
