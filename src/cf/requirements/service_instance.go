package requirements

import (
	"cf"
	"cf/api"
	"cf/configuration"
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
	config          *configuration.Configuration
	serviceRepo     api.ServiceRepository
	serviceInstance cf.ServiceInstance
}

func NewServiceInstanceRequirement(name string, ui terminal.UI, config *configuration.Configuration, sR api.ServiceRepository) (req *ServiceInstanceApiRequirement) {
	req = new(ServiceInstanceApiRequirement)
	req.name = name
	req.ui = ui
	req.config = config
	req.serviceRepo = sR
	return
}

func (req *ServiceInstanceApiRequirement) Execute() (success bool) {
	var apiErr *net.ApiError
	req.serviceInstance, apiErr = req.serviceRepo.FindInstanceByName(req.name)

	// todo - this seems like a special case; confirm?
	if !req.serviceInstance.IsFound() {
		req.ui.Failed("Service %s does not exist.", req.name)
		return false
	}

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *ServiceInstanceApiRequirement) GetServiceInstance() cf.ServiceInstance {
	return req.serviceInstance
}
