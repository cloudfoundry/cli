package requirements

import (
	"cf/api"
	"cf/errors"
	"cf/models"
	"cf/terminal"
)

type ApplicationRequirement interface {
	Requirement
	GetApplication() models.Application
}

type applicationApiRequirement struct {
	name        string
	ui          terminal.UI
	appRepo     api.ApplicationRepository
	application models.Application
}

func NewApplicationRequirement(name string, ui terminal.UI, aR api.ApplicationRepository) (req *applicationApiRequirement) {
	req = new(applicationApiRequirement)
	req.name = name
	req.ui = ui
	req.appRepo = aR
	return
}

func (req *applicationApiRequirement) Execute() (success bool) {
	var apiResponse errors.Error
	req.application, apiResponse = req.appRepo.Read(req.name)

	if apiResponse != nil {
		req.ui.Failed(apiResponse.Error())
		return false
	}

	return true
}

func (req *applicationApiRequirement) GetApplication() models.Application {
	return req.application
}
