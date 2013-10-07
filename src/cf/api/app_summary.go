package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strconv"
)

type AppSummaryRepository interface {
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

func (repo CloudControllerAppSummaryRepository) GetSummary(app cf.Application) (summary cf.AppSummary, apiResponse net.ApiResponse) {
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
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	statsResponse := StatsApiResponse{}

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &statsResponse)
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
