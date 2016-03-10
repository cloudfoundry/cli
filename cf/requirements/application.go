package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter -o fakes/fake_application_requirement.go . ApplicationRequirement
type ApplicationRequirement interface {
	Requirement
	GetApplication() models.Application
}

type applicationApiRequirement struct {
	name        string
	appRepo     applications.ApplicationRepository
	application models.Application
}

func NewApplicationRequirement(name string, aR applications.ApplicationRepository) *applicationApiRequirement {
	req := &applicationApiRequirement{}
	req.name = name
	req.appRepo = aR
	return req
}

func (req *applicationApiRequirement) Execute() error {
	var apiErr error
	req.application, apiErr = req.appRepo.Read(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *applicationApiRequirement) GetApplication() models.Application {
	return req.application
}
