package commands

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ConfigCommands struct {
	ui     terminal.UI
	config configuration.ReadWriter
}

func NewConfig(ui terminal.UI, config configuration.ReadWriter) ConfigCommands {
	return ConfigCommands{ui: ui, config: config}
}

func (cmd ConfigCommands) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "config",
		Description: "write default values to the config",
		Usage:       "CF_NAME config [--async-timeout TIMEOUT_IN_MINUTES] [--trace true | false | path/to/file]",
		Flags: []cli.Flag{
			flag_helpers.NewIntFlag("async-timeout", "Timeout for async HTTP requests"),
			flag_helpers.NewStringFlag("trace", "Trace HTTP requests"),
		},
	}
}

func (cmd ConfigCommands) GetRequirements(_ requirements.Factory, _ *cli.Context) ([]requirements.Requirement, error) {
	return nil, nil
}

func (cmd ConfigCommands) Run(context *cli.Context) {
	if !context.IsSet("trace") && !context.IsSet("async-timeout") {
		cmd.ui.FailWithUsage(context)
		return
	}

	if context.IsSet("async-timeout") {
		asyncTimeout := context.Int("async-timeout")
		if asyncTimeout < 0 {
			cmd.ui.FailWithUsage(context)
		}

		cmd.config.SetAsyncTimeout(uint(asyncTimeout))
	}

	if context.IsSet("trace") {
		cmd.config.SetTrace(context.String("trace"))
	}
}
