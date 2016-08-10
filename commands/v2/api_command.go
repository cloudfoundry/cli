package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type ApiCommand struct {
	OptionalArgs      flags.APITarget `positional-args:"yes"`
	Unset             bool            `long:"unset" description:"Remove all api endpoint targeting"`
	SkipSSLValidation bool            `long:"skip-ssl-validation" description:"Skip verification of the API endpoint. Not recommended!"`
	usage             interface{}     `usage:"CF_NAME api [URL]"`
}

func (_ ApiCommand) Setup() error {
	return nil
}

func (_ ApiCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
