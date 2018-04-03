package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

//go:generate counterfeiter . LoginStrategy

type LoginStrategy interface {
	Try() error
}

type LogInCmd struct {
	basicStrategy LoginStrategy
	uaaStrategy   LoginStrategy
	director      boshdir.Director
}

func NewLogInCmd(
	basicStrategy LoginStrategy,
	uaaStrategy LoginStrategy,
	director boshdir.Director,
) LogInCmd {
	return LogInCmd{
		basicStrategy: basicStrategy,
		uaaStrategy:   uaaStrategy,
		director:      director,
	}
}

func (c LogInCmd) Run() error {
	info, err := c.director.Info()
	if err != nil {
		return err
	}

	switch info.Auth.Type {
	case "uaa":
		return c.uaaStrategy.Try()
	case "basic":
		return c.basicStrategy.Try()
	default:
		return bosherr.Errorf("Unknown auth type '%s'", info.Auth.Type)
	}
}
