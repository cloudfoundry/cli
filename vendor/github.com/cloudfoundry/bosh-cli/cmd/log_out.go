package cmd

import (
	"errors"

	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type LogOutCmd struct {
	environment string
	config      cmdconf.Config
	ui          biui.UI
}

func NewLogOutCmd(environment string, config cmdconf.Config, ui biui.UI) LogOutCmd {
	return LogOutCmd{environment: environment, config: config, ui: ui}
}

func (c LogOutCmd) Run() error {
	if c.environment == "" {
		return errors.New("Expected non-empty Director URL")
	}

	updatedConfig := c.config.UnsetCredentials(c.environment)

	err := updatedConfig.Save()
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Logged out from '%s'", c.environment)

	return nil
}
