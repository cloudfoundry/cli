package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter -o fakes/fake_application_requirement.go . ApplicationRequirement
type ApplicationRequirement interface {
	Requirement
	GetApplication() models.Application
}

type applicationApiRequirement struct {
	name        string
	ui          terminal.UI
	appRepo     applications.ApplicationRepository
	application models.Application
}

func NewApplicationRequirement(name string, ui terminal.UI, aR applications.ApplicationRepository) *applicationApiRequirement {
	req := &applicationApiRequirement{}
	req.name = name
	req.ui = ui
	req.appRepo = aR
	return req
}

func (req *applicationApiRequirement) Execute() (success bool) {
	var apiErr error
	req.application, apiErr = req.appRepo.Read(req.name)

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *applicationApiRequirement) GetApplication() models.Application {
	return req.application
}
