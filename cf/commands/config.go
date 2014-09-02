package commands

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
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
		Description: T("write default values to the config"),
		Usage:       T("CF_NAME config [--async-timeout TIMEOUT_IN_MINUTES] [--trace true | false | path/to/file] [--color true | false] [--locale (LOCALE | CLEAR)]"),
		Flags: []cli.Flag{
			flag_helpers.NewIntFlag("async-timeout", T("Timeout for async HTTP requests")),
			flag_helpers.NewStringFlag("trace", T("Trace HTTP requests")),
			flag_helpers.NewStringFlag("color", T("Enable or disable color")),
			flag_helpers.NewStringFlag("locale", "Set default locale. If LOCALE is CLEAR, previous locale is deleted."),
		},
	}
}

func (cmd ConfigCommands) GetRequirements(_ requirements.Factory, _ *cli.Context) ([]requirements.Requirement, error) {
	return nil, nil
}

func (cmd ConfigCommands) Run(context *cli.Context) {
	if !context.IsSet("trace") && !context.IsSet("async-timeout") && !context.IsSet("color") && !context.IsSet("locale") {
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

	if context.IsSet("color") {
		value := context.String("color")
		switch value {
		case "true":
			cmd.config.SetColorEnabled("true")
		case "false":
			cmd.config.SetColorEnabled("false")
		default:
			cmd.ui.FailWithUsage(context)
		}
	}

	if context.IsSet("locale") {
		locale := context.String("locale")

		if locale == "CLEAR" {
			cmd.config.SetLocale("")
			return
		}

		foundLocale := false

		for _, value := range SUPPORTED_LOCALES {
			if value == locale {
				cmd.config.SetLocale(locale)
				foundLocale = true
				break
			}
		}

		if !foundLocale {
			cmd.ui.Say(fmt.Sprintf("Could not find locale %s. The known locales are:", locale))
			cmd.ui.Say("")
			for _, locale := range SUPPORTED_LOCALES {
				cmd.ui.Say(locale)
			}
		}
	}
}
