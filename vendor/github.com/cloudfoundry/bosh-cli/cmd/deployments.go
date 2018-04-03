package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeploymentsCmd struct {
	ui       biui.UI
	director boshdir.Director
}

func NewDeploymentsCmd(ui biui.UI, director boshdir.Director) DeploymentsCmd {
	return DeploymentsCmd{ui: ui, director: director}
}

func (c DeploymentsCmd) Run() error {
	deployments, err := c.director.Deployments()
	if err != nil {
		return err
	}

	return DeploymentsTable{deployments, c.ui}.Print()
}
