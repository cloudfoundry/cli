package v2

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type ConfigCommand struct {
	AsyncTimeout int               `long:"async-timeout" description:"Timeout for async HTTP requests"`
	Color        flag.Color        `long:"color" description:"Enable or disable color"`
	Locale       flag.Locale       `long:"locale" description:"Set default locale. If LOCALE is 'CLEAR', previous locale is deleted."`
	Trace        flag.PathWithBool `long:"trace" description:"Trace HTTP requests"`
	usage        interface{}       `usage:"CF_NAME config [--async-timeout TIMEOUT_IN_MINUTES] [--trace (true | false | path/to/file)] [--color (true | false)] [--locale (LOCALE | CLEAR)]"`
}

func (ConfigCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (ConfigCommand) Execute(args []string) error {
	return translatableerror.UnrefactoredCommandError{}
}
