package requirements

import "code.cloudfoundry.org/cli/cf/configuration/coreconfig"

type configRefreshingRequirement struct {
	requirement     Requirement
	configRefresher ConfigRefresher
}

//go:generate counterfeiter . ConfigRefresher

type ConfigRefresher interface {
	Refresh() (coreconfig.Warning, error)
}

func NewConfigRefreshingRequirement(requirement Requirement, configRefresher ConfigRefresher) configRefreshingRequirement {
	return configRefreshingRequirement{
		requirement:     requirement,
		configRefresher: configRefresher,
	}
}

func (c configRefreshingRequirement) Execute() error {
	err := c.requirement.Execute()
	if err != nil {
		// Do the config refresh
		_, err = c.configRefresher.Refresh()
		if err != nil {
			return err
		}

		return c.requirement.Execute()
	}

	return nil
}
