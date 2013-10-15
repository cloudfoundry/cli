package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
)

type ServiceInstanceRequirement interface {
	Requirement
	GetServiceInstance() cf.ServiceInstance
}

type serviceInstanceApiRequirement struct {
	name            string
	ui              terminal.UI
	serviceRepo     api.ServiceRepository
	serviceInstance cf.ServiceInstance
}

func newServiceInstanceRequirement(name string, ui terminal.UI, sR api.ServiceRepository) (req *serviceInstanceApiRequirement) {
	req = new(serviceInstanceApiRequirement)
	req.name = name
	req.ui = ui
	req.serviceRepo = sR
	return
}

func (req *serviceInstanceApiRequirement) Execute() (success bool) {
	var apiResponse net.ApiResponse
	req.serviceInstance, apiResponse = req.serviceRepo.FindInstanceByName(req.name)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *serviceInstanceApiRequirement) GetServiceInstance() cf.ServiceInstance {
	return req.serviceInstance
}
