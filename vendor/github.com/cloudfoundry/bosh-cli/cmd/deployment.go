package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeploymentCmd struct {
	sessionFactory func(cmdconf.Config) Session

	config cmdconf.Config
	ui     biui.UI
}

func NewDeploymentCmd(
	sessionFactory func(cmdconf.Config) Session,
	config cmdconf.Config,
	ui biui.UI,
) DeploymentCmd {
	return DeploymentCmd{sessionFactory: sessionFactory, config: config, ui: ui}
}

func (c DeploymentCmd) Run() error {
	sess := c.sessionFactory(c.config)

	deployment, err := sess.Deployment()
	if err != nil {
		return err
	}

	return DeploymentsTable{[]boshdir.Deployment{deployment}, c.ui}.Print()
}
