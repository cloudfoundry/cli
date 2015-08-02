package command_factory

import (
	"errors"

	"github.com/cloudfoundry/cli/plugin/rpc"

	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/codegangsta/cli"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type Factory interface {
	GetByCmdName(cmdName string) (cmd command.Command, err error)
	CommandMetadatas() []command_metadata.CommandMetadata
	CheckIfCoreCmdExists(cmdName string) bool
	GetCommandFlags(string) []string
	GetCommandTotalArgs(string) (int, error)
}

type concreteFactory struct {
	cmdsByName map[string]command.Command
}

func NewFactory(ui terminal.UI, config core_config.ReadWriter, manifestRepo manifest.ManifestRepository, repoLocator api.RepositoryLocator, pluginConfig plugin_config.PluginConfiguration, rpcService *rpc.CliRpcService) (factory concreteFactory) {
	factory.cmdsByName = make(map[string]command.Command)

	factory.cmdsByName["create-user"] = user.NewCreateUser(ui, config, repoLocator.GetUserRepository())
	factory.cmdsByName["restart-app-instance"] = application.NewRestartAppInstance(ui, config, repoLocator.GetAppInstancesRepository())

	return
}

func (f concreteFactory) GetByCmdName(cmdName string) (cmd command.Command, err error) {
	cmd, found := f.cmdsByName[cmdName]
	if !found {
		for _, c := range f.cmdsByName {
			if c.Metadata().ShortName == cmdName {
				return c, nil
			}
		}

		err = errors.New(T("Command not found"))
	}
	return
}

func (f concreteFactory) CheckIfCoreCmdExists(cmdName string) bool {
	if _, exists := f.cmdsByName[cmdName]; exists {
		return true
	}

	for _, singleCmd := range f.cmdsByName {
		metaData := singleCmd.Metadata()
		if metaData.ShortName == cmdName {
			return true
		}
	}

	return false
}

func (factory concreteFactory) CommandMetadatas() (commands []command_metadata.CommandMetadata) {
	for _, command := range factory.cmdsByName {
		commands = append(commands, command.Metadata())
	}
	return
}

func (f concreteFactory) GetCommandFlags(command string) []string {
	cmd, err := f.GetByCmdName(command)
	if err != nil {
		return []string{}
	}

	var flags []string
	for _, flag := range cmd.Metadata().Flags {
		switch t := flag.(type) {
		default:
		case flag_helpers.StringSliceFlagWithNoDefault:
			flags = append(flags, t.Name)
		case flag_helpers.IntFlagWithNoDefault:
			flags = append(flags, t.Name)
		case flag_helpers.StringFlagWithNoDefault:
			flags = append(flags, t.Name)
		case cli.BoolFlag:
			flags = append(flags, t.Name)
		}
	}

	return flags
}

func (f concreteFactory) GetCommandTotalArgs(command string) (int, error) {
	cmd, err := f.GetByCmdName(command)
	if err != nil {
		return 0, err
	}

	return cmd.Metadata().TotalArgs, nil
}
