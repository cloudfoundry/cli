package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

type ConfigCommand struct {
	AsyncTimeout int    `long:"async-timeout" description:"Timeout for async HTTP requests"`
	Trace        string `long:"trace" description:"Trace HTTP requests"`
	Color        string `long:"color" description:"Enable or disable color"`
	Locale       string `long:"locale" description:"Set default locale. If LOCALE is 'CLEAR', previous locale is deleted."`
}

func (_ ConfigCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
