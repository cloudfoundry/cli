package v7

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ConfigCommand struct {
	UI           command.UI
	Config       command.Config
	AsyncTimeout flag.Timeout      `long:"async-timeout" description:"Timeout in minutes for async HTTP requests"`
	Color        flag.Color        `long:"color" description:"Enable or disable color in CLI output"`
	Locale       flag.Locale       `long:"locale" description:"Set default locale. If LOCALE is 'CLEAR', previous locale is deleted."`
	Trace        flag.PathWithBool `long:"trace" description:"Trace HTTP requests by default. If a file path is provided then output will write to the file provided. If the file does not exist it will be created."`
	usage        interface{}       `usage:"CF_NAME config [--async-timeout TIMEOUT_IN_MINUTES] [--trace (true | false | path/to/file)] [--color (true | false)] [--locale (LOCALE | CLEAR)]"`
}

func (cmd *ConfigCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui

	return nil
}

func (cmd ConfigCommand) Execute(args []string) error {
	if !cmd.Color.IsSet && cmd.Trace == "" && cmd.Locale.Locale == "" && !cmd.AsyncTimeout.IsSet {
		return translatableerror.IncorrectUsageError{Message: "at least one flag must be provided"}
	}

	cmd.UI.DisplayText("Setting values in config...")

	if cmd.AsyncTimeout.IsSet {
		cmd.Config.SetAsyncTimeout(cmd.AsyncTimeout.Value)
	}

	if cmd.Color.IsSet {
		cmd.Config.SetColorEnabled(cmd.Color.Value)
	}

	if cmd.Locale.Locale != "" {
		cmd.Config.SetLocale(cmd.Locale.Locale)
	}

	if cmd.Trace != "" {
		cmd.Config.SetTrace(string(cmd.Trace))
	}

	cmd.UI.DisplayOK()
	return nil
}
