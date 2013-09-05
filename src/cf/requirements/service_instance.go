package requirements

import (
	"cf"
	"cf/api"
	"cf/configuration"
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

func (req *ServiceInstanceApiRequirement) Execute() (err error) {
	req.serviceInstance, err = req.serviceRepo.FindInstanceByName(req.config, req.name)
	if err != nil {
		req.ui.Failed("", err)
	}
	return
}

func (req *ServiceInstanceApiRequirement) GetServiceInstance() cf.ServiceInstance {
	return req.serviceInstance
}
