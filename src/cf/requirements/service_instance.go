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

type ServiceInstanceApiRequirement struct {
	name            string
	ui              terminal.UI
	serviceRepo     api.ServiceRepository
	serviceInstance cf.ServiceInstance
}

func NewServiceInstanceRequirement(name string, ui terminal.UI, sR api.ServiceRepository) (req *ServiceInstanceApiRequirement) {
	req = new(ServiceInstanceApiRequirement)
	req.name = name
	req.ui = ui
	req.serviceRepo = sR
	return
}

func (req *ServiceInstanceApiRequirement) Execute() (success bool) {
	var apiStatus net.ApiStatus
	req.serviceInstance, apiStatus = req.serviceRepo.FindInstanceByName(req.name)

	if apiStatus.NotSuccessful() {
		req.ui.Failed(apiStatus.Message)
		return false
	}

	return true
}

func (req *ServiceInstanceApiRequirement) GetServiceInstance() cf.ServiceInstance {
	return req.serviceInstance
}
