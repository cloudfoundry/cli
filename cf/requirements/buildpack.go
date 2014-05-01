package requirements

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type BuildpackRequirement interface {
	Requirement
	GetBuildpack() models.Buildpack
}

type buildpackApiRequirement struct {
	name          string
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
	buildpack     models.Buildpack
}

func NewBuildpackRequirement(name string, ui terminal.UI, bR api.BuildpackRepository) (req *buildpackApiRequirement) {
	req = new(buildpackApiRequirement)
	req.name = name
	req.ui = ui
	req.buildpackRepo = bR
	return
}

func (req *buildpackApiRequirement) Execute() (success bool) {
	var apiErr error
	req.buildpack, apiErr = req.buildpackRepo.FindByName(req.name)

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *buildpackApiRequirement) GetBuildpack() models.Buildpack {
	return req.buildpack
}
