package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type CurlCommand struct {
	RequiredArgs          flag.APIPath `positional-args:"yes"`
	CustomHeaders         []string     `short:"H" description:"Custom headers to include in the request, flag can be specified multiple times"`
	HTTPMethod            string       `short:"X" description:"HTTP method (GET,POST,PUT,DELETE,etc)"`
	HTTPData              string       `short:"d" description:"HTTP data to include in the request body, or '@' followed by a file name to read the data from"`
	IncludeReponseHeaders bool         `short:"i" description:"Include response headers in the output"`
	OutputFile            string       `long:"output" description:"Write curl body to FILE instead of stdout"`
	usage                 interface{}  `usage:"CF_NAME curl PATH [-iv] [-X METHOD] [-H HEADER] [-d DATA] [--output FILE]\n\n   By default 'CF_NAME curl' will perform a GET to the specified PATH. If data\n   is provided via -d, a POST will be performed instead, and the Content-Type\n   will be set to application/json. You may override headers with -H and the\n   request method with -X.\n\n   For API documentation, please visit http://apidocs.cloudfoundry.org.\n\nEXAMPLES:\n   CF_NAME curl \"/v2/apps\" -X GET -H \"Content-Type: application/x-www-form-urlencoded\" -d 'q=name:myapp'\n   CF_NAME curl \"/v2/apps\" -d @/path/to/file"`
}

func (_ CurlCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (_ CurlCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
