package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type SpaceRequirement interface {
	Requirement
	GetSpace() models.Space
}

type spaceApiRequirement struct {
	name      string
	ui        terminal.UI
	spaceRepo spaces.SpaceRepository
	space     models.Space
}

func NewSpaceRequirement(name string, ui terminal.UI, sR spaces.SpaceRepository) (req *spaceApiRequirement) {
	req = new(spaceApiRequirement)
	req.name = name
	req.ui = ui
	req.spaceRepo = sR
	return
}

func (req *spaceApiRequirement) Execute() (success bool) {
	var apiErr error
	req.space, apiErr = req.spaceRepo.FindByName(req.name)

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *spaceApiRequirement) GetSpace() models.Space {
	return req.space
}
