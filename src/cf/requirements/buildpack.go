package requirements

import (
	"cf/api"
	"cf/models"
	"cf/net"
	"cf/terminal"
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
	var apiResponse net.ApiResponse
	req.buildpack, apiResponse = req.buildpackRepo.FindByName(req.name)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *buildpackApiRequirement) GetBuildpack() models.Buildpack {
	return req.buildpack
}
