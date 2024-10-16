package requirements

import (
	"code.cloudfoundry.org/cli/v8/cf/api"
	"code.cloudfoundry.org/cli/v8/cf/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceInstanceRequirement

type ServiceInstanceRequirement interface {
	Requirement
	GetServiceInstance() models.ServiceInstance
}

type serviceInstanceAPIRequirement struct {
	name            string
	serviceRepo     api.ServiceRepository
	serviceInstance models.ServiceInstance
}

func NewServiceInstanceRequirement(name string, sR api.ServiceRepository) (req *serviceInstanceAPIRequirement) {
	req = new(serviceInstanceAPIRequirement)
	req.name = name
	req.serviceRepo = sR
	return
}

func (req *serviceInstanceAPIRequirement) Execute() error {
	var apiErr error
	req.serviceInstance, apiErr = req.serviceRepo.FindInstanceByName(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *serviceInstanceAPIRequirement) GetServiceInstance() models.ServiceInstance {
	return req.serviceInstance
}
