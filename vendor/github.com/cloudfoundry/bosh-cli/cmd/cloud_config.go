package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type CloudConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewCloudConfigCmd(ui boshui.UI, director boshdir.Director) CloudConfigCmd {
	return CloudConfigCmd{ui: ui, director: director}
}

func (c CloudConfigCmd) Run() error {
	cloudConfig, err := c.director.LatestCloudConfig()
	if err != nil {
		return err
	}

	c.ui.PrintBlock([]byte(cloudConfig.Properties))

	return nil
}
