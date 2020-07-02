package v7

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strings"

	"fmt"

	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type CurlCommand struct {
	BaseCommand

	RequiredArgs          flag.APIPath    `positional-args:"yes"`
	CustomHeaders         []string        `short:"H" description:"Custom headers to include in the request, flag can be specified multiple times"`
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
	err := mergeHeaders(&header, cmd.CustomHeaders)
	if err != nil {
		return translatableerror.RequestCreationError{Err: err}
	}

	method := cmd.HTTPMethod
	if cmd.HTTPMethod == "" && cmd.HTTPData != "" {
		method = "POST"
	}

	byteString := []byte(cmd.HTTPData)
	trimmedInput := strings.Trim(string(cmd.HTTPData), `"'`)
	if strings.HasPrefix(trimmedInput, `@`) {
		trimmedInput = strings.Trim(trimmedInput[1:], `"'`)
		byteString, err = ioutil.ReadFile(trimmedInput)
		if err != nil {
			return translatableerror.RequestCreationError{Err: err}
		}
	}

	responseBodyBytes, httpResponse, err := cmd.cloudControllerClient.MakeRequestSendReceiveRaw(
		method,
		url,
		header,
		byteString,
	)

	if err != nil && cmd.FailOnHTTPError {
		return translatableerror.CurlExit22Error{StatusCode: httpResponse.StatusCode}
	}

	var bytesToWrite []byte

	if cmd.IncludeReponseHeaders {
		headerBytes, _ := httputil.DumpResponse(httpResponse, false)
		bytesToWrite = append(bytesToWrite, headerBytes...)
	}

	bytesToWrite = append(bytesToWrite, responseBodyBytes...)

	if cmd.OutputFile != "" {
		err = ioutil.WriteFile(cmd.OutputFile.String(), bytesToWrite, 0666)
		if err != nil {
			// Todo: change this error type
			return translatableerror.ManifestCreationError{Err: err}
		}

		cmd.UI.DisplayOK()
	} else {
		cmd.UI.DisplayText(string(bytesToWrite))
	}

	return nil
}

func mergeHeaders(destination *http.Header, customHeaders []string) (err error) {
	headerString := strings.Join(customHeaders, "\n")
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
			destination.Add(key, value)
		}
	}

	return
}
