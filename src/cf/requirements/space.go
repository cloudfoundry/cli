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
	var apiErr *net.ApiError
	req.space, apiErr = req.spaceRepo.FindByName(req.name)

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *SpaceApiRequirement) GetSpace() cf.Space {
	return req.space
}
