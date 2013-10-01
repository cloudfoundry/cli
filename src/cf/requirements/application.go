package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
)

type ApplicationRequirement interface {
	Requirement
	GetApplication() cf.Application
}

type ApplicationApiRequirement struct {
	name        string
	ui          terminal.UI
	appRepo     api.ApplicationRepository
	application cf.Application
}

func NewApplicationRequirement(name string, ui terminal.UI, aR api.ApplicationRepository) (req *ApplicationApiRequirement) {
	req = new(ApplicationApiRequirement)
	req.name = name
	req.ui = ui
	req.appRepo = aR
	return
}

func (req *ApplicationApiRequirement) Execute() (success bool) {
	var apiStatus net.ApiStatus
	req.application, apiStatus = req.appRepo.FindByName(req.name)

	if apiStatus.IsError() {
		req.ui.Failed(apiStatus.Message)
		return false
	}

	return req.application.IsFound()
}

func (req *ApplicationApiRequirement) GetApplication() cf.Application {
	return req.application
}
