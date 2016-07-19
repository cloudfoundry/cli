package commands

import (
	"errors"
	"sort"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type ConfigCommands struct {
	ui     terminal.UI
	config coreconfig.ReadWriter
}

func init() {
	commandregistry.Register(&ConfigCommands{})
}

func (cmd *ConfigCommands) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["async-timeout"] = &flags.IntFlag{Name: "async-timeout", Usage: T("Timeout for async HTTP requests")}
	fs["trace"] = &flags.StringFlag{Name: "trace", Usage: T("Trace HTTP requests")}
	fs["color"] = &flags.StringFlag{Name: "color", Usage: T("Enable or disable color")}
	fs["locale"] = &flags.StringFlag{Name: "locale", Usage: T("Set default locale. If LOCALE is 'CLEAR', previous locale is deleted.")}

	return commandregistry.CommandMetadata{
		Name:        "config",
		Description: T("Write default values to the config"),
		Usage: []string{
			T("CF_NAME config [--async-timeout TIMEOUT_IN_MINUTES] [--trace (true | false | path/to/file)] [--color (true | false)] [--locale (LOCALE | CLEAR)]"),
		},
		Flags: fs,
	}
}

func (cmd *ConfigCommands) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd *ConfigCommands) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	return cmd
}

func (cmd *ConfigCommands) Execute(context flags.FlagContext) error {
	if !context.IsSet("trace") && !context.IsSet("async-timeout") && !context.IsSet("color") && !context.IsSet("locale") {
		return errors.New(T("Incorrect Usage") + "\n\n" + commandregistry.Commands.CommandUsage("config"))
	}

	if context.IsSet("async-timeout") {
		asyncTimeout := context.Int("async-timeout")
		if asyncTimeout < 0 {
			return errors.New(T("Incorrect Usage") + "\n\n" + commandregistry.Commands.CommandUsage("config"))
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
			return errors.New(T("Incorrect Usage") + "\n\n" + commandregistry.Commands.CommandUsage("config"))
		}
	}

	if context.IsSet("locale") {
		locale := context.String("locale")

		if locale == "CLEAR" {
			cmd.config.SetLocale("")
			return nil
		}

		if IsSupportedLocale(locale) {
			cmd.config.SetLocale(locale)
			return nil
		}

		unsupportedLocaleMessage := T("Could not find locale '{{.UnsupportedLocale}}'. The known locales are:\n", map[string]interface{}{
			"UnsupportedLocale": locale,
		})
		supportedLocales := SupportedLocales()
		sort.Strings(supportedLocales)
		for i := range supportedLocales {
			unsupportedLocaleMessage = unsupportedLocaleMessage + "\n" + supportedLocales[i]
		}

		return errors.New(unsupportedLocaleMessage)
	}
	return nil
}
