package app_instances

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State   string
	Since   float64
	Details string
}

type StatsApiResponse map[string]InstanceStatsApiResponse

type InstanceStatsApiResponse struct {
	Stats struct {
		DiskQuota int64 `json:"disk_quota"`
		MemQuota  int64 `json:"mem_quota"`
		Usage     struct {
			Cpu  float64
			Disk int64
			Mem  int64
		}
	}
}

//go:generate counterfeiter -o fakes/fake_app_instances_repository.go . AppInstancesRepository
type AppInstancesRepository interface {
	GetInstances(appGuid string) (instances []models.AppInstanceFields, apiErr error)
	DeleteInstance(appGuid string, instance int) error
}

type CloudControllerAppInstancesRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerAppInstancesRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerAppInstancesRepository) {
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
			State:   models.InstanceState(strings.ToLower(v.State)),
			Details: v.Details,
			Since:   time.Unix(int64(v.Since), 0),
		}
	}

	return repo.updateInstancesWithStats(appGuid, instances)
}

func (repo CloudControllerAppInstancesRepository) DeleteInstance(appGuid string, instance int) error {
	err := repo.gateway.DeleteResource(repo.config.ApiEndpoint(), fmt.Sprintf("/v2/apps/%s/instances/%d", appGuid, instance))
	if err != nil {
		return err
	}
	return nil
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
