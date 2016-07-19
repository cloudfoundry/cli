package requirements

import (
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/models"
)

//go:generate counterfeiter . UserRequirement

type UserRequirement interface {
	Requirement
	GetUser() models.UserFields
}

type userAPIRequirement struct {
	username string
	userRepo api.UserRepository
	wantGUID bool

	user models.UserFields
}

func NewUserRequirement(
	username string,
	userRepo api.UserRepository,
	wantGUID bool,
) *userAPIRequirement {
	req := new(userAPIRequirement)
	req.username = username
	req.userRepo = userRepo
	req.wantGUID = wantGUID

	return req
}

func (req *userAPIRequirement) Execute() error {
	if req.wantGUID {
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

func (req *userAPIRequirement) GetUser() models.UserFields {
	return req.user
}
