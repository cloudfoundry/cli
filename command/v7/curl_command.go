package v7

import (
	"bufio"
	"net/http"
	"net/textproto"
	"strings"

	"fmt"

	"code.cloudfoundry.org/cli/command/flag"
)

type CurlCommand struct {
	BaseCommand

	RequiredArgs          flag.APIPath    `positional-args:"yes"`
	CustomHeaders         []string          `short:"H" description:"Custom headers to include in the request, flag can be specified multiple times"`
	HTTPMethod            string          `short:"X" description:"HTTP method (GET,POST,PUT,DELETE,etc)"`
	HTTPData              flag.PathWithAt `short:"d" description:"HTTP data to include in the request body, or '@' followed by a file name to read the data from"`
	FailOnHTTPError       bool            `short:"f" long:"fail" description:"Server errors return exit code 22"`
	IncludeReponseHeaders bool            `short:"i" description:"Include response headers in the output"`
	OutputFile            flag.Path       `long:"output" description:"Write curl body to FILE instead of stdout"`
	usage                 interface{}     `usage:"CF_NAME curl PATH [-iv] [-X METHOD] [-H HEADER]... [-d DATA] [--output FILE]\n\n   By default 'CF_NAME curl' will perform a GET to the specified PATH. If data\n   is provided via -d, a POST will be performed instead, and the Content-Type\n   will be set to application/json. You may override headers with -H and the\n   request method with -X.\n\n   For API documentation, please visit http://apidocs.cloudfoundry.org.\n\nEXAMPLES:\n   CF_NAME curl \"/v2/apps\" -X GET -H \"Content-Type: application/x-www-form-urlencoded\" -d 'q=name:myapp'\n   CF_NAME curl \"/v2/apps\" -d @/path/to/file"`
}

func (cmd CurlCommand) Execute(args []string) error {
	url := fmt.Sprintf("%s/%s", cmd.Config.Target(), strings.TrimLeft(cmd.RequiredArgs.Path, "/"))
	header := http.Header{}
	err := mergeHeaders(header, strings.Join(cmd.CustomHeaders, "\n"))
	if err != nil {
		// err = fmt.Errorf("%s: %s", T("Error parsing headers"), err.Error())
		return err
	}

	method := cmd.HTTPMethod
	if cmd.HTTPMethod == "" && cmd.HTTPData != "" {
		method = "POST"
	}

	responseBytes, warnings, err := cmd.cloudControllerClient.MakeRequestSendReceiveRaw(
		method,
		url,
		header,
		[]byte(cmd.HTTPData),
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(string(responseBytes))

	return nil
}

func mergeHeaders(destination http.Header, headerString string) (err error) {
	headerString = strings.TrimSpace(headerString)
	headerString += "\n\n"
	headerReader := bufio.NewReader(strings.NewReader(headerString))
	headers, err := textproto.NewReader(headerReader).ReadMIMEHeader()
	if err != nil {
		return
	}

	for key, values := range headers {
		destination.Del(key)
		for _, value := range values {
			fmt.Printf("adding key %v and value %v\n\n", key, value)
			destination.Add(key, value)
		}
	}

	return
}
