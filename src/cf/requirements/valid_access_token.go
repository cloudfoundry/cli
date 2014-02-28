package requirements

import (
	"cf/api"
	"cf/terminal"
)

type ValidAccessTokenRequirement struct {
	ui      terminal.UI
	appRepo api.ApplicationRepository
}

func NewValidAccessTokenRequirement(ui terminal.UI, appRepo api.ApplicationRepository) ValidAccessTokenRequirement {
	return ValidAccessTokenRequirement{ui, appRepo}
}

func (req ValidAccessTokenRequirement) Execute() (success bool) {
	_, apiResponse := req.appRepo.Read("checking_for_valid_access_token")

	if apiResponse != nil && apiResponse.StatusCode() == 401 {
		req.ui.Say(terminal.NotLoggedInText())
		return false
	}

	return true
}
