package commands

import (
	"sort"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ConfigCommands struct {
	ui     terminal.UI
	config core_config.ReadWriter
}

func init() {
	command_registry.Register(&ConfigCommands{})
}

func (cmd *ConfigCommands) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["async-timeout"] = &flags.IntFlag{Name: "async-timeout", Usage: T("Timeout for async HTTP requests")}
	fs["trace"] = &flags.StringFlag{Name: "trace", Usage: T("Trace HTTP requests")}
	fs["color"] = &flags.StringFlag{Name: "color", Usage: T("Enable or disable color")}
	fs["locale"] = &flags.StringFlag{Name: "locale", Usage: T("Set default locale. If LOCALE is 'CLEAR', previous locale is deleted.")}

	return command_registry.CommandMetadata{
		Name:        "config",
		Description: T("Write default values to the config"),
		Usage: []string{
			T("CF_NAME config [--async-timeout TIMEOUT_IN_MINUTES] [--trace (true | false | path/to/file)] [--color (true | false)] [--locale (LOCALE | CLEAR)]"),
		},
		Flags: fs,
	}
}

func (cmd *ConfigCommands) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *ConfigCommands) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	return cmd
}

func (cmd *ConfigCommands) Execute(context flags.FlagContext) {
	if !context.IsSet("trace") && !context.IsSet("async-timeout") && !context.IsSet("color") && !context.IsSet("locale") {
		cmd.ui.Failed(T("Incorrect Usage") + "\n\n" + command_registry.Commands.CommandUsage("config"))
		return
	}

	if context.IsSet("async-timeout") {
		asyncTimeout := context.Int("async-timeout")
		if asyncTimeout < 0 {
			cmd.ui.Failed(T("Incorrect Usage") + "\n\n" + command_registry.Commands.CommandUsage("config"))
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
			cmd.ui.Failed(T("Incorrect Usage") + "\n\n" + command_registry.Commands.CommandUsage("config"))
		}
	}

	if context.IsSet("locale") {
		locale := context.String("locale")

		if locale == "CLEAR" {
			cmd.config.SetLocale("")
			return
		}

		if IsSupportedLocale(locale) {
			cmd.config.SetLocale(locale)
			return
		}

		unsupportedLocaleMessage := T("Could not find locale '{{.UnsupportedLocale}}'. The known locales are:\n", map[string]interface{}{
			"UnsupportedLocale": locale,
		})
		supportedLocales := SupportedLocales()
		sort.Strings(supportedLocales)
		for i := range supportedLocales {
			unsupportedLocaleMessage = unsupportedLocaleMessage + "\n" + supportedLocales[i]
		}

		cmd.ui.Failed(unsupportedLocaleMessage)
	}
}
