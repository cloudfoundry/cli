package requirements

import (
	"cf/api"
	"cf/errors"
	"cf/models"
	"cf/terminal"
)

type ServiceInstanceRequirement interface {
	Requirement
	GetServiceInstance() models.ServiceInstance
}

type serviceInstanceApiRequirement struct {
	name            string
	ui              terminal.UI
	serviceRepo     api.ServiceRepository
	serviceInstance models.ServiceInstance
}

func NewServiceInstanceRequirement(name string, ui terminal.UI, sR api.ServiceRepository) (req *serviceInstanceApiRequirement) {
	req = new(serviceInstanceApiRequirement)
	req.name = name
	req.ui = ui
	req.serviceRepo = sR
	return
}

func (req *serviceInstanceApiRequirement) Execute() (success bool) {
	var apiResponse errors.Error
	req.serviceInstance, apiResponse = req.serviceRepo.FindInstanceByName(req.name)

	if apiResponse != nil {
		req.ui.Failed(apiResponse.Error())
		return false
	}

	return true
}

func (req *serviceInstanceApiRequirement) GetServiceInstance() models.ServiceInstance {
	return req.serviceInstance
}
