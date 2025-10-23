package requirements

import (
	"code.cloudfoundry.org/cli/v8/cf/api/spaces"
	"code.cloudfoundry.org/cli/v8/cf/models"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SpaceRequirement

type SpaceRequirement interface {
	Requirement
	SetSpaceName(string)
	GetSpace() models.Space
}

type spaceAPIRequirement struct {
	name      string
	spaceRepo spaces.SpaceRepository
	space     models.Space
}

func NewSpaceRequirement(name string, sR spaces.SpaceRepository) *spaceAPIRequirement {
	req := &spaceAPIRequirement{}
	req.name = name
	req.spaceRepo = sR
	return req
}

func (req *spaceAPIRequirement) SetSpaceName(name string) {
	req.name = name
}

func (req *spaceAPIRequirement) Execute() error {
	var apiErr error
	req.space, apiErr = req.spaceRepo.FindByName(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *spaceAPIRequirement) GetSpace() models.Space {
	return req.space
}
