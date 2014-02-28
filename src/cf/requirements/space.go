package requirements

import (
	"cf/api"
	"cf/errors"
	"cf/models"
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
	var apiResponse errors.Error
	req.space, apiResponse = req.spaceRepo.FindByName(req.name)

	if apiResponse != nil {
		req.ui.Failed(apiResponse.Error())
		return false
	}

	return true
}

func (req *spaceApiRequirement) GetSpace() models.Space {
	return req.space
}
