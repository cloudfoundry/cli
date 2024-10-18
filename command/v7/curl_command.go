package v7

import (
	"net/http/httputil"
	"os"

	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
)

type CurlCommand struct {
	BaseCommand

	RequiredArgs           flag.APIPath    `positional-args:"yes"`
	CustomHeaders          []string        `short:"H" description:"Custom headers to include in the request, flag can be specified multiple times"`
	HTTPMethod             string          `short:"X" description:"HTTP method (GET,POST,PUT,DELETE,etc)"`
	HTTPData               flag.PathWithAt `short:"d" description:"HTTP data to include in the request body, or '@' followed by a file name to read the data from"`
	FailOnHTTPError        bool            `short:"f" long:"fail" description:"Server errors return exit code 22"`
	IncludeResponseHeaders bool            `short:"i" description:"Include response headers in the output"`
	OutputFile             flag.Path       `long:"output" description:"Write curl body to FILE instead of stdout"`
	usage                  interface{}     `usage:"CF_NAME curl PATH [-iv] [-X METHOD] [-H HEADER]... [-d DATA] [--output FILE]\n\n   By default 'CF_NAME curl' will perform a GET to the specified PATH. If data\n   is provided via -d, a POST will be performed instead, and the Content-Type\n   will be set to application/json. You may override headers with -H and the\n   request method with -X.\n\n   For API documentation, please visit http://v2-apidocs.cloudfoundry.org.\n\nEXAMPLES:\n   CF_NAME curl \"/v2/apps\" -X GET -H \"Content-Type: application/x-www-form-urlencoded\" -d 'q=name:myapp'\n   CF_NAME curl \"/v2/apps\" -d @/path/to/file"`
}

func (cmd CurlCommand) Execute(args []string) error {
	responseBodyBytes, httpResponse, err := cmd.Actor.MakeCurlRequest(
		cmd.HTTPMethod,
		cmd.RequiredArgs.Path,
		cmd.CustomHeaders,
		string(cmd.HTTPData),
		cmd.FailOnHTTPError,
	)

	if err != nil {
		return err
	}

	if alreadyWroteVerboseOutput, _ := cmd.Config.Verbose(); alreadyWroteVerboseOutput {
		return nil
	}

	var bytesToWrite []byte

	if cmd.IncludeResponseHeaders {
		headerBytes, _ := httputil.DumpResponse(httpResponse, false)
		bytesToWrite = append(bytesToWrite, headerBytes...)
	}

	bytesToWrite = append(bytesToWrite, responseBodyBytes...)

	if cmd.OutputFile != "" {
		err = os.WriteFile(cmd.OutputFile.String(), bytesToWrite, 0666)
		if err != nil {
			return translatableerror.FileCreationError{Err: err}
		}

		cmd.UI.DisplayOK()
	} else {
		cmd.UI.DisplayText(string(bytesToWrite))
	}

	return nil
}
