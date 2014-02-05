package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
)

type UserRequirement interface {
	Requirement
	GetUser() cf.UserFields
}

type userApiRequirement struct {
	username string
	ui       terminal.UI
	userRepo api.UserRepository
	user     cf.UserFields
}

func NewUserRequirement(username string, ui terminal.UI, userRepo api.UserRepository) (req *userApiRequirement) {
	req = new(userApiRequirement)
	req.username = username
	req.ui = ui
	req.userRepo = userRepo
	return
}

func (req *userApiRequirement) Execute() (success bool) {
	var apiResponse net.ApiResponse
	req.user, apiResponse = req.userRepo.FindByUsername(req.username)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *userApiRequirement) GetUser() cf.UserFields {
	return req.user
}
