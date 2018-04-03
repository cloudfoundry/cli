package cmd

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type UAALoginStrategy struct {
	sessionFactory func(cmdconf.Config) Session

	config cmdconf.Config
	ui     boshui.UI

	logTag string
	logger boshlog.Logger

	successMsg string
	failureMsg string
}

func NewUAALoginStrategy(
	sessionFactory func(cmdconf.Config) Session,
	config cmdconf.Config,
	ui boshui.UI,
	logger boshlog.Logger,
) UAALoginStrategy {
	return UAALoginStrategy{
		sessionFactory: sessionFactory,
		config:         config,
		ui:             ui,

		logTag: "UAALoginStrategy",
		logger: logger,

		successMsg: "Successfully authenticated with UAA",
		failureMsg: "Failed to authenticate with UAA",
	}
}

func (c UAALoginStrategy) Try() error {
	sess := c.sessionFactory(c.config)

	uaa, err := sess.UAA()
	if err != nil {
		return err
	}

	if sess.Credentials().IsUAAClient() {
		return c.tryClient(uaa) // Only try once
	} else {
		return c.tryUser(sess, uaa)
	}
}

func (c UAALoginStrategy) tryClient(uaa boshuaa.UAA) error {
	_, err := uaa.ClientCredentialsGrant()

	if err == nil {
		c.ui.PrintLinef(c.successMsg)
	} else {
		c.ui.ErrorLinef(c.failureMsg)
	}

	// Dont bother saving client token since there is no way to refresh it
	return err
}

func (c UAALoginStrategy) tryUser(sess Session, uaa boshuaa.UAA) error {
	prompts, err := uaa.Prompts()
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Using environment '%s'", sess.Environment())

	for {
		authed, err := c.tryUserOnce(sess.Environment(), prompts, uaa)
		if err != nil {
			return err
		}

		if authed {
			return nil
		}
	}
}

func (c UAALoginStrategy) tryUserOnce(environment string, prompts []boshuaa.Prompt, uaa boshuaa.UAA) (bool, error) {
	var answers []boshuaa.PromptAnswer

	for _, prompt := range prompts {
		var askFunc func(string) (string, error)

		if prompt.IsPassword() {
			askFunc = c.ui.AskForPassword
		} else {
			askFunc = c.ui.AskForText
		}

		value, err := askFunc(prompt.Label)
		if err != nil {
			return false, err
		}

		if value != "" {
			answer := boshuaa.PromptAnswer{Key: prompt.Key, Value: value}
			answers = append(answers, answer)
		}
	}

	accessToken, err := uaa.OwnerPasswordCredentialsGrant(answers)
	if err != nil {
		c.logger.Error(c.logTag, "Failed to get access token: %s", err)
		c.ui.ErrorLinef(c.failureMsg)
		return false, nil
	}

	creds := cmdconf.Creds{
		RefreshToken: accessToken.RefreshToken().Value(),
	}

	updatedConfig := c.config.SetCredentials(environment, creds)

	err = updatedConfig.Save()
	if err != nil {
		return false, err
	}

	c.ui.PrintLinef(c.successMsg)

	return true, nil
}
