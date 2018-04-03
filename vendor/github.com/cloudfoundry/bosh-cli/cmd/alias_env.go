package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type AliasEnvCmd struct {
	sessionFactory func(cmdconf.Config) Session

	config cmdconf.Config
	ui     boshui.UI
}

func NewAliasEnvCmd(
	sessionFactory func(cmdconf.Config) Session,
	config cmdconf.Config,
	ui boshui.UI,
) AliasEnvCmd {
	return AliasEnvCmd{sessionFactory: sessionFactory, config: config, ui: ui}
}

func (c AliasEnvCmd) Run(opts AliasEnvOpts) error {
	updatedConfig, err := c.config.AliasEnvironment(opts.URL, opts.Args.Alias, opts.CACert.Content)
	if err != nil {
		return err
	}

	sess := c.sessionFactory(updatedConfig)

	director, err := sess.Director()
	if err != nil {
		return err
	}

	info, err := director.Info()
	if err != nil {
		return err
	}

	err = updatedConfig.Save()
	if err != nil {
		return err
	}

	InfoTable{info, c.ui}.Print()

	return nil
}
