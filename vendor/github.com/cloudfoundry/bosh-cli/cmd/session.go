package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshuit "github.com/cloudfoundry/bosh-cli/ui/task"
)

type SessionImpl struct {
	context SessionContext

	ui               boshui.UI
	printEnvironment bool
	printDeployment  bool

	logger boshlog.Logger

	// Memoized
	director        boshdir.Director
	directorInfo    boshdir.Info
	directorInfoSet bool
}

func NewSessionImpl(
	context SessionContext,
	ui boshui.UI,
	printEnvironment bool,
	printDeployment bool,
	logger boshlog.Logger,
) *SessionImpl {
	return &SessionImpl{
		context: context,

		ui:               ui,
		printEnvironment: printEnvironment,
		printDeployment:  printDeployment,

		logger: logger,
	}
}

func (c SessionImpl) Environment() string        { return c.context.Environment() }
func (c SessionImpl) Credentials() cmdconf.Creds { return c.context.Credentials() }

func (c SessionImpl) UAA() (boshuaa.UAA, error) {
	err := c.setDirectorInfo()
	if err != nil {
		return nil, err
	}

	uaaURL := c.directorInfo.Auth.Options["url"]

	uaaURLStr, ok := uaaURL.(string)
	if !ok {
		return nil, bosherr.Errorf("Expected URL '%s' to be a string", uaaURL)
	}

	uaaConfig, err := boshuaa.NewConfigFromURL(uaaURLStr)
	if err != nil {
		return nil, err
	}

	uaaConfig.CACert = c.context.CACert()

	creds := c.Credentials()
	uaaConfig.Client = creds.Client
	uaaConfig.ClientSecret = creds.ClientSecret

	if len(uaaConfig.Client) == 0 {
		uaaConfig.Client = "bosh_cli"
	}

	return boshuaa.NewFactory(c.logger).New(uaaConfig)
}

func (c *SessionImpl) Director() (boshdir.Director, error) {
	if c.director != nil {
		return c.director, nil
	}

	dirConfig, err := boshdir.NewConfigFromURL(c.Environment())
	if err != nil {
		return nil, err
	}

	dirConfig.CACert = c.context.CACert()

	creds := c.Credentials()

	err = c.setDirectorInfo()
	if err != nil {
		return nil, err
	}

	if c.directorInfo.Auth.Type != "uaa" {
		dirConfig.Client = creds.Client
		dirConfig.ClientSecret = creds.ClientSecret
	} else if creds.IsUAA() {
		uaa, err := c.UAA()
		if err != nil {
			return nil, err
		}

		if creds.IsUAAClient() {
			dirConfig.TokenFunc = boshuaa.NewClientTokenSession(uaa).TokenFunc
		} else {
			origToken := uaa.NewStaleAccessToken(creds.RefreshToken)
			dirConfig.TokenFunc = boshuaa.NewAccessTokenSession(origToken).TokenFunc
		}
	}

	if c.printEnvironment {
		c.ui.PrintLinef("Using environment '%s' as %s", c.Environment(), creds.Description())
	}

	taskReporter := boshuit.NewReporter(c.ui, true)
	fileReporter := boshui.NewFileReporter(c.ui)

	director, err := boshdir.NewFactory(c.logger).New(dirConfig, taskReporter, fileReporter)
	if err != nil {
		return nil, err
	}

	// Memoize only on successfuly creation
	c.director = director

	return c.director, nil
}

func (c *SessionImpl) setDirectorInfo() error {
	if c.directorInfoSet {
		return nil
	}

	director, err := c.AnonymousDirector()
	if err != nil {
		return err
	}

	info, err := director.Info()
	if err != nil {
		return err
	}

	c.directorInfo = info
	c.directorInfoSet = true
	return nil
}

func (c *SessionImpl) AnonymousDirector() (boshdir.Director, error) {
	dirConfig, err := boshdir.NewConfigFromURL(c.Environment())
	if err != nil {
		return nil, err
	}

	dirConfig.CACert = c.context.CACert()

	return boshdir.NewFactory(c.logger).New(dirConfig, nil, nil)
}

func (c *SessionImpl) Deployment() (boshdir.Deployment, error) {
	director, err := c.Director()
	if err != nil {
		return nil, err
	}

	deploymentName := c.context.Deployment()

	deployment, err := director.FindDeployment(deploymentName)
	if err != nil {
		return nil, err
	}

	if c.printDeployment {
		c.ui.PrintLinef("Using deployment '%s'", deploymentName)
	}

	return deployment, nil
}
