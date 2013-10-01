package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strconv"
)

type AppSummaryRepository interface {
	GetSummary(app cf.Application) (summary cf.AppSummary, apiStatus net.ApiStatus)
}

type CloudControllerAppSummaryRepository struct {
	config  configuration.Configuration
	gateway net.Gateway
	appRepo ApplicationRepository
}

func NewCloudControllerAppSummaryRepository(config configuration.Configuration, gateway net.Gateway, appRepo ApplicationRepository) (repo CloudControllerAppSummaryRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.appRepo = appRepo
	return
}

func (repo CloudControllerAppSummaryRepository) GetSummary(app cf.Application) (summary cf.AppSummary, apiStatus net.ApiStatus) {
	summary.App = app

	instances, apiStatus := repo.appRepo.GetInstances(app)
	if apiStatus.IsError() {
		return
	}

	instances, apiStatus = repo.updateInstancesWithStats(app, instances)
	if apiStatus.IsError() {
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

func (repo CloudControllerAppSummaryRepository) updateInstancesWithStats(app cf.Application, instances []cf.ApplicationInstance) (updatedInst []cf.ApplicationInstance, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", repo.config.Target, app.Guid)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	apiResponse := StatsApiResponse{}

	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, &apiResponse)
	if apiStatus.IsError() {
		return
	}

	updatedInst = make([]cf.ApplicationInstance, len(apiResponse), len(apiResponse))
	for k, v := range apiResponse {
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
