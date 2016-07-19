package requirements

import (
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/models"
)

//go:generate counterfeiter . ApplicationRequirement

type ApplicationRequirement interface {
	Requirement
	GetApplication() models.Application
}

type applicationAPIRequirement struct {
	name        string
	appRepo     applications.Repository
	application models.Application
}

func NewApplicationRequirement(name string, aR applications.Repository) *applicationAPIRequirement {
	req := &applicationAPIRequirement{}
	req.name = name
	req.appRepo = aR
	return req
}

func (req *applicationAPIRequirement) Execute() error {
	var apiErr error
	req.application, apiErr = req.appRepo.Read(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *applicationAPIRequirement) GetApplication() models.Application {
	return req.application
}
