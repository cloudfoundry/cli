package requirements

import (
	"cf/terminal"
	"cf/configuration"
	"errors"
)

type LoginRequirement struct {
	ui terminal.UI
	config *configuration.Configuration
}

func NewLoginRequirement(ui terminal.UI, config *configuration.Configuration) LoginRequirement {
	return LoginRequirement{ui, config}
}

func (req LoginRequirement) Execute() (err error) {
	if !req.config.IsLoggedIn() {
		req.ui.Say("Not logged in. Use '%s' to log in.", terminal.Yellow("cf login"))
		err = errors.New("You need to be logged in")
		return
	}
	return
}
