package rpc

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
)

//go:generate counterfeiter . CommandRunner

type CommandRunner interface {
	Command([]string, commandregistry.Dependency, bool) error
}

type commandRunner struct{}

func NewCommandRunner() CommandRunner {
	return &commandRunner{}
}

func (c *commandRunner) Command(args []string, deps commandregistry.Dependency, pluginApiCall bool) (err error) {
	cmdRegistry := commandregistry.Commands

	if cmdRegistry.CommandExists(args[0]) {
		fc := flags.NewFlagContext(cmdRegistry.FindCommand(args[0]).MetaData().Flags)
		err = fc.Parse(args[1:]...)
		if err != nil {
			return err
		}

		cfCmd := cmdRegistry.FindCommand(args[0])
		cfCmd = cfCmd.SetDependency(deps, pluginApiCall)

		reqs, reqErr := cfCmd.Requirements(requirements.NewFactory(deps.Config, deps.RepoLocator), fc)
		if reqErr != nil {
			return reqErr
		}

		for _, r := range reqs {
			if err = r.Execute(); err != nil {
				return err
			}
		}

		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("command panic: %v", r)
			}
		}()

		return cfCmd.Execute(fc)
	}

	return nil
}
