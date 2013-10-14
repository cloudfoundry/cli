package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
)

type UserRequirement interface {
	Requirement
	GetUser() cf.User
}

type UserApiRequirement struct {
	username string
	ui       terminal.UI
	userRepo api.UserRepository
	user     cf.User
}

func NewUserRequirement(username string, ui terminal.UI, userRepo api.UserRepository) (req *UserApiRequirement) {
	req = new(UserApiRequirement)
	req.username = username
	req.ui = ui
	req.userRepo = userRepo
	return
}

func (req *UserApiRequirement) Execute() (success bool) {
	var apiResponse net.ApiResponse
	req.user, apiResponse = req.userRepo.FindByUsername(req.username)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *UserApiRequirement) GetUser() cf.User {
	return req.user
}
