package requirements

import (
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/models"
)

//go:generate counterfeiter . BuildpackRequirement

type BuildpackRequirement interface {
	Requirement
	GetBuildpack() models.Buildpack
}

type buildpackAPIRequirement struct {
	name          string
	stack         string
	buildpackRepo api.BuildpackRepository
	buildpack     models.Buildpack
}

func NewBuildpackRequirement(name, stack string, bR api.BuildpackRepository) (req *buildpackAPIRequirement) {
	req = new(buildpackAPIRequirement)
	req.name = name
	req.stack = stack
	req.buildpackRepo = bR
	return
}

func (req *buildpackAPIRequirement) Execute() error {
	var apiErr error
	if req.stack == "" {
		req.buildpack, apiErr = req.buildpackRepo.FindByName(req.name)
	} else {
		req.buildpack, apiErr = req.buildpackRepo.FindByNameAndStack(req.name, req.stack)
	}

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *buildpackAPIRequirement) GetBuildpack() models.Buildpack {
	return req.buildpack
}
