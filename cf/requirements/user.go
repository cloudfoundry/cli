package requirements

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter -o fakes/fake_user_requirement.go . UserRequirement
type UserRequirement interface {
	Requirement
	GetUser() models.UserFields
}

type userApiRequirement struct {
	username string
	userRepo api.UserRepository
	wantGuid bool

	user models.UserFields
}

func NewUserRequirement(
	username string,
	userRepo api.UserRepository,
	wantGuid bool,
) *userApiRequirement {
	req := new(userApiRequirement)
	req.username = username
	req.userRepo = userRepo
	req.wantGuid = wantGuid

	return req
}

func (req *userApiRequirement) Execute() error {
	if req.wantGuid {
		var err error
		req.user, err = req.userRepo.FindByUsername(req.username)
		if err != nil {
			return err
		}
	} else {
		req.user = models.UserFields{Username: req.username}
	}

	return nil
}

func (req *userApiRequirement) GetUser() models.UserFields {
	return req.user
}
