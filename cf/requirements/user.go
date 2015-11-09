package requirements

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/feature_flags"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type UserRequirement interface {
	Requirement
	GetUser() models.UserFields
}

type userApiRequirement struct {
	username string
	ui       terminal.UI
	userRepo api.UserRepository
	flagRepo feature_flags.FeatureFlagRepository
	config   core_config.Reader

	user models.UserFields
}

func NewUserRequirement(
	username string,
	ui terminal.UI,
	userRepo api.UserRepository,
	flagRepo feature_flags.FeatureFlagRepository,
	config core_config.Reader,
) *userApiRequirement {
	req := new(userApiRequirement)
	req.username = username
	req.ui = ui
	req.userRepo = userRepo
	req.config = config
	req.flagRepo = flagRepo

	return req
}

func (req *userApiRequirement) Execute() bool {
	if req.config.IsMinApiVersion("2.37.0") {
		setRolesByUsernameFlag, err := req.flagRepo.FindByName("set_roles_by_username")
		if err == nil && setRolesByUsernameFlag.Enabled {
			req.user = models.UserFields{Username: req.username}
			return true
		}
	}

	var err error
	req.user, err = req.userRepo.FindByUsername(req.username)
	if err != nil {
		req.ui.Failed(err.Error())
		return false
	}

	return true
}

func (req *userApiRequirement) GetUser() models.UserFields {
	return req.user
}
