package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
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
	GetInstances(appGuid string) (instances []models.AppInstanceFields, apiErr error)
}

type CloudControllerAppInstancesRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerAppInstancesRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerAppInstancesRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppInstancesRepository) GetInstances(appGuid string) (instances []models.AppInstanceFields, err error) {
	instancesResponse := InstancesApiResponse{}
	err = repo.gateway.GetResource(
		fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.ApiEndpoint(), appGuid),
		&instancesResponse)
	if err != nil {
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

func (repo CloudControllerAppInstancesRepository) updateInstancesWithStats(guid string, instances []models.AppInstanceFields) (updatedInst []models.AppInstanceFields, apiErr error) {
	path := fmt.Sprintf("%s/v2/apps/%s/stats", repo.config.ApiEndpoint(), guid)
	statsResponse := StatsApiResponse{}
	apiErr = repo.gateway.GetResource(path, &statsResponse)
	if apiErr != nil {
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
