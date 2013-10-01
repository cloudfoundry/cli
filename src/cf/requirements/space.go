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

type SpaceApiRequirement struct {
	name      string
	ui        terminal.UI
	spaceRepo api.SpaceRepository
	space     cf.Space
}

func NewSpaceRequirement(name string, ui terminal.UI, sR api.SpaceRepository) (req *SpaceApiRequirement) {
	req = new(SpaceApiRequirement)
	req.name = name
	req.ui = ui
	req.spaceRepo = sR
	return
}

func (req *SpaceApiRequirement) Execute() (success bool) {
	var apiStatus net.ApiStatus
	req.space, apiStatus = req.spaceRepo.FindByName(req.name)

	if apiStatus.IsError() {
		req.ui.Failed(apiStatus.Message)
		return false
	}

	return req.space.IsFound()
}

func (req *SpaceApiRequirement) GetSpace() cf.Space {
	return req.space
}
