package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
)

type ConfigCommand struct {
	AsyncTimeout int         `long:"async-timeout" description:"Timeout for async HTTP requests"`
	Color        string      `long:"color" description:"Enable or disable color"`
	Locale       string      `long:"locale" description:"Set default locale. If LOCALE is 'CLEAR', previous locale is deleted."`
	Trace        string      `long:"trace" description:"Trace HTTP requests"`
	usage        interface{} `usage:"CF_NAME config [--async-timeout TIMEOUT_IN_MINUTES] [--trace (true | false | path/to/file)] [--color (true | false)] [--locale (LOCALE | CLEAR)]"`
}

func (_ ConfigCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ ConfigCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
