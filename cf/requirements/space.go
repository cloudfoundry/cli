package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter -o fakes/fake_space_requirement.go . SpaceRequirement
type SpaceRequirement interface {
	Requirement
	SetSpaceName(string)
	GetSpace() models.Space
}

type spaceApiRequirement struct {
	name      string
	spaceRepo spaces.SpaceRepository
	space     models.Space
}

func NewSpaceRequirement(name string, sR spaces.SpaceRepository) *spaceApiRequirement {
	req := &spaceApiRequirement{}
	req.name = name
	req.spaceRepo = sR
	return req
}

func (req *spaceApiRequirement) SetSpaceName(name string) {
	req.name = name
}

func (req *spaceApiRequirement) Execute() error {
	var apiErr error
	req.space, apiErr = req.spaceRepo.FindByName(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *spaceApiRequirement) GetSpace() models.Space {
	return req.space
}
