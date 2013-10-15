package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
)

type SpaceRequirement interface {
	Requirement
	GetSpace() cf.Space
}

type spaceApiRequirement struct {
	name      string
	ui        terminal.UI
	spaceRepo api.SpaceRepository
	space     cf.Space
}

func newSpaceRequirement(name string, ui terminal.UI, sR api.SpaceRepository) (req *spaceApiRequirement) {
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

func (req *spaceApiRequirement) GetSpace() cf.Space {
	return req.space
}
