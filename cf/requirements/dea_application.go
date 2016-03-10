package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/errors"
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

func NewDEAApplicationRequirement(name string, applicationRepo applications.ApplicationRepository) DEAApplicationRequirement {
	return &deaApplicationRequirement{
		appName: name,
		appRepo: applicationRepo,
	}
}

func (req *deaApplicationRequirement) Execute() error {
	app, err := req.appRepo.Read(req.appName)
	if err != nil {
		return err
	}

	if app.Diego == true {
		return errors.New("The app is running on the Diego backend, which does not support this command.")
	}

	req.application = app

	return nil
}

func (req *deaApplicationRequirement) GetApplication() models.Application {
	return req.application
}
