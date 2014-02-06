package requirements

import (
	"cf/api"
	"cf/models"
	"cf/net"
	"cf/terminal"
)

type SpaceRequirement interface {
	Requirement
	GetSpace() models.Space
}

type spaceApiRequirement struct {
	name      string
	ui        terminal.UI
	spaceRepo api.SpaceRepository
	space     models.Space
}

func NewSpaceRequirement(name string, ui terminal.UI, sR api.SpaceRepository) (req *spaceApiRequirement) {
	req = new(spaceApiRequirement)
	req.name = name
	req.ui = ui
	req.spaceRepo = sR
	return
}

func (req *spaceApiRequirement) Execute() (success bool) {
	var apiResponse net.ApiResponse
	req.space, apiResponse = req.spaceRepo.FindByName(req.name)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *spaceApiRequirement) GetSpace() models.Space {
	return req.space
}
