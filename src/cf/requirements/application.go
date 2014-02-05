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

type applicationApiRequirement struct {
	name        string
	ui          terminal.UI
	appRepo     api.ApplicationRepository
	application cf.Application
}

func NewApplicationRequirement(name string, ui terminal.UI, aR api.ApplicationRepository) (req *applicationApiRequirement) {
	req = new(applicationApiRequirement)
	req.name = name
	req.ui = ui
	req.appRepo = aR
	return
}

func (req *applicationApiRequirement) Execute() (success bool) {
	var apiResponse net.ApiResponse
	req.application, apiResponse = req.appRepo.Read(req.name)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *applicationApiRequirement) GetApplication() cf.Application {
	return req.application
}
