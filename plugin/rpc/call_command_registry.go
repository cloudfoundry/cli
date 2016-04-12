package rpc

import (
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
)

//go:generate counterfeiter . CommandRunner

type CommandRunner interface {
	Command([]string, commandregistry.Dependency, bool) error
}

type commandRunner struct{}

func NewCommandRunner() CommandRunner {
	return &commandRunner{}
}

func (c *commandRunner) Command(args []string, deps commandregistry.Dependency, pluginApiCall bool) error {
	var err error

	cmdRegistry := commandregistry.Commands

	if cmdRegistry.CommandExists(args[0]) {
		fc := flags.NewFlagContext(cmdRegistry.FindCommand(args[0]).MetaData().Flags)
		err = fc.Parse(args[1:]...)
		if err != nil {
			return err
		}

		cfCmd := cmdRegistry.FindCommand(args[0])
		cfCmd = cfCmd.SetDependency(deps, pluginApiCall)

		reqs := cfCmd.Requirements(requirements.NewFactory(deps.Config, deps.RepoLocator), fc)

		for _, r := range reqs {
			if err = r.Execute(); err != nil {
				return err
			}
		}

		cfCmd.Execute(fc)
	}

	return nil
}
