package requirements

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/terminal"
)

type ApplicationRequirement struct {
	name    string
	ui      terminal.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository

	Application cf.Application
}

func NewApplicationRequirement(name string, ui terminal.UI, config *configuration.Configuration, aR api.ApplicationRepository) (req ApplicationRequirement) {
	req.name = name
	req.ui = ui
	req.config = config
	req.appRepo = aR
	return
}

func (req *ApplicationRequirement) Execute() (err error) {
	req.Application, err = req.appRepo.FindByName(req.config, req.name)
	if err != nil {
		req.ui.Failed("", err)
	}
	return
}
