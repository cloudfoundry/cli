package requirements

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter -o fakes/fake_service_instance_requirement.go . ServiceInstanceRequirement
type ServiceInstanceRequirement interface {
	Requirement
	GetServiceInstance() models.ServiceInstance
}

type serviceInstanceApiRequirement struct {
	name            string
	serviceRepo     api.ServiceRepository
	serviceInstance models.ServiceInstance
}

func NewServiceInstanceRequirement(name string, sR api.ServiceRepository) (req *serviceInstanceApiRequirement) {
	req = new(serviceInstanceApiRequirement)
	req.name = name
	req.serviceRepo = sR
	return
}

func (req *serviceInstanceApiRequirement) Execute() error {
	var apiErr error
	req.serviceInstance, apiErr = req.serviceRepo.FindInstanceByName(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *serviceInstanceApiRequirement) GetServiceInstance() models.ServiceInstance {
	return req.serviceInstance
}
