package requirements

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

type BuildpackRequirement interface {
	Requirement
	GetBuildpack() models.Buildpack
}

type buildpackApiRequirement struct {
	name          string
	buildpackRepo api.BuildpackRepository
	buildpack     models.Buildpack
}

func NewBuildpackRequirement(name string, bR api.BuildpackRepository) (req *buildpackApiRequirement) {
	req = new(buildpackApiRequirement)
	req.name = name
	req.buildpackRepo = bR
	return
}

func (req *buildpackApiRequirement) Execute() error {
	var apiErr error
	req.buildpack, apiErr = req.buildpackRepo.FindByName(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *buildpackApiRequirement) GetBuildpack() models.Buildpack {
	return req.buildpack
}
