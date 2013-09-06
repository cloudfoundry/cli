package requirements

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/terminal"
)

type ApplicationRequirement interface {
	Requirement
	GetApplication() cf.Application
}

type ApplicationApiRequirement struct {
	name        string
	ui          terminal.UI
	config      *configuration.Configuration
	appRepo     api.ApplicationRepository
	application cf.Application
}

func NewApplicationRequirement(name string, ui terminal.UI, config *configuration.Configuration, aR api.ApplicationRepository) (req *ApplicationApiRequirement) {
	req = new(ApplicationApiRequirement)
	req.name = name
	req.ui = ui
	req.config = config
	req.appRepo = aR
	return
}

func (req *ApplicationApiRequirement) Execute() (err error) {
	req.application, err = req.appRepo.FindByName(req.name)
	if err != nil {
		req.ui.Failed("", err)
	}
	return
}

func (req *ApplicationApiRequirement) GetApplication() cf.Application {
	return req.application
}
