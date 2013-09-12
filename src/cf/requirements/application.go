package requirements

import (
	"cf"
	"cf/api"
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
	var apiErr *api.ApiError
	req.application, apiErr = req.appRepo.FindByName(req.name)

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *ApplicationApiRequirement) GetApplication() cf.Application {
	return req.application
}
