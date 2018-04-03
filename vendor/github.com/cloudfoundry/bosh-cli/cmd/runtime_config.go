package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type RuntimeConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewRuntimeConfigCmd(ui boshui.UI, director boshdir.Director) RuntimeConfigCmd {
	return RuntimeConfigCmd{ui: ui, director: director}
}

func (c RuntimeConfigCmd) Run(opts RuntimeConfigOpts) error {
	runtimeConfig, err := c.director.LatestRuntimeConfig(opts.Name)
	if err != nil {
		return err
	}

	c.ui.PrintBlock([]byte(runtimeConfig.Properties))

	return nil
}
