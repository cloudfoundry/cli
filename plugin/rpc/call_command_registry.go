package rpc

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
)

type NonCodegangstaRunner interface {
	Command([]string, command_registry.Dependency, bool) error
}

type nonCodegangstaRunner struct{}

func NewNonCodegangstaRunner() NonCodegangstaRunner {
	return &nonCodegangstaRunner{}
}

func (c *nonCodegangstaRunner) Command(args []string, deps command_registry.Dependency, pluginApiCall bool) error {
	var err error

	cmdRegistry := command_registry.Commands

	if cmdRegistry.CommandExists(args[0]) {
		fc := flags.NewFlagContext(cmdRegistry.FindCommand(args[0]).MetaData().Flags)
		err = fc.Parse(args[1:]...)
		if err != nil {
			return err
		}

		cfCmd := cmdRegistry.FindCommand(args[0])
		cfCmd = cfCmd.SetDependency(deps, pluginApiCall)

		reqs, err := cfCmd.Requirements(requirements.NewFactory(deps.Ui, deps.Config, deps.RepoLocator), fc)
		if err != nil {
			return err
		}

		for _, r := range reqs {
			if !r.Execute() {
				return errors.New("Error in requirement")
			}
		}

		cfCmd.Execute(fc)
	}

	return nil
}
