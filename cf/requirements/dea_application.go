package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter -o fakes/fake_dea_application_requirement.go . DEAApplicationRequirement
type DEAApplicationRequirement interface {
	Requirement
	GetApplication() models.Application
}

type deaApplicationRequirement struct {
	appName string
	ui      terminal.UI
	appRepo applications.ApplicationRepository

	application models.Application
}

func NewDEAApplicationRequirement(name string, ui terminal.UI, applicationRepo applications.ApplicationRepository) DEAApplicationRequirement {
	return &deaApplicationRequirement{
		appName: name,
		ui:      ui,
		appRepo: applicationRepo,
	}
}

func (req *deaApplicationRequirement) Execute() (success bool) {
	app, err := req.appRepo.Read(req.appName)
	if err != nil {
		req.ui.Failed(err.Error())
		return false
	}

	if app.Diego == true {
		req.ui.Failed("The app is running on the Diego backend, which does not support this command.")
		return false
	}

	req.application = app

	return true
}

func (req *deaApplicationRequirement) GetApplication() models.Application {
	return req.application
}
