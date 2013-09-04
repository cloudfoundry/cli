package requirements

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/terminal"
)

type ServiceInstanceRequirement struct {
	name            string
	ui              terminal.UI
	config          *configuration.Configuration
	serviceRepo     api.ServiceRepository
	ServiceInstance cf.ServiceInstance
}

func NewServiceInstanceRequirement(name string, ui terminal.UI, config *configuration.Configuration, sR api.ServiceRepository) (req ServiceInstanceRequirement) {
	req.name = name
	req.ui = ui
	req.config = config
	req.serviceRepo = sR
	return
}

func (req *ServiceInstanceRequirement) Execute() (err error) {
	req.ServiceInstance, err = req.serviceRepo.FindInstanceByName(req.config, req.name)
	if err != nil {
		req.ui.Failed("", err)
	}
	return
}
