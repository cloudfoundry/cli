package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands/flags"
)

type CurlCommand struct {
	RequiredArgs          flags.APIPath `positional-args:"yes"`
	IncludeReponseHeaders bool          `short:"i" description:"Include response headers in the output"`
	HTTPMethod            string        `short:"X" description:"HTTP method (GET,POST,PUT,DELETE,etc)"`
	CustomHeaders         []string      `short:"H" description:"Custom headers to include in the request, flag can be specified multiple times"`
	HTTPData              string        `short:"d" description:"HTTP data to include in the request body, or '@' followed by a file name to read the data from"`
	OutputFile            string        `long:"output" description:"Write curl body to FILE instead of stdout"`
}

func (_ CurlCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
