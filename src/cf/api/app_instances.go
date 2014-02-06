package api

import (
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
	Since float64
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

type AppInstancesRepository interface {
	GetInstances(appGuid string) (instances []models.AppInstanceFields, apiResponse net.ApiResponse)
}

type CloudControllerAppInstancesRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerAppInstancesRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerAppInstancesRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppInstancesRepository) GetInstances(appGuid string) (instances []models.AppInstanceFields, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.Target, appGuid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instancesResponse := InstancesApiResponse{}

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &instancesResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instances = make([]models.AppInstanceFields, len(instancesResponse), len(instancesResponse))
	for k, v := range instancesResponse {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}

		instances[index] = models.AppInstanceFields{
			State: models.InstanceState(strings.ToLower(v.State)),
			Since: time.Unix(int64(v.Since), 0),
		}
	}

	return repo.updateInstancesWithStats(appGuid, instances)
}

func (repo CloudControllerAppInstancesRepository) updateInstancesWithStats(guid string, instances []models.AppInstanceFields) (updatedInst []models.AppInstanceFields, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", repo.config.Target, guid)
	statsResponse := StatsApiResponse{}
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, &statsResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedInst = make([]models.AppInstanceFields, len(statsResponse), len(statsResponse))
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
