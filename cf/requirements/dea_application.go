package requirements

import (
	"code.cloudfoundry.org/cli/v8/cf/api/applications"
	"code.cloudfoundry.org/cli/v8/cf/errors"
	"code.cloudfoundry.org/cli/v8/cf/models"
	"code.cloudfoundry.org/cli/v8/cf/terminal"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DEAApplicationRequirement

type DEAApplicationRequirement interface {
	Requirement
	GetApplication() models.Application
}

type deaApplicationRequirement struct {
	appName string
	ui      terminal.UI
	appRepo applications.Repository

	application models.Application
}

func NewDEAApplicationRequirement(name string, applicationRepo applications.Repository) DEAApplicationRequirement {
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
