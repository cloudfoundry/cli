package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/util"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
)

type Curl struct {
	ui       terminal.UI
	config   coreconfig.Reader
	curlRepo api.CurlRepository
}

func init() {
	commandregistry.Register(&Curl{})
}

func (cmd *Curl) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &flags.BoolFlag{ShortName: "i", Usage: T("Include response headers in the output")}
	fs["X"] = &flags.StringFlag{ShortName: "X", Usage: T("HTTP method (GET,POST,PUT,DELETE,etc)")}
	fs["H"] = &flags.StringSliceFlag{ShortName: "H", Usage: T("Custom headers to include in the request, flag can be specified multiple times")}
	fs["d"] = &flags.StringFlag{ShortName: "d", Usage: T("HTTP data to include in the request body, or '@' followed by a file name to read the data from")}
	fs["output"] = &flags.StringFlag{Name: "output", Usage: T("Write curl body to FILE instead of stdout")}

	return commandregistry.CommandMetadata{
		Name:        "curl",
		Description: T("Executes a request to the targeted API endpoint"),
		Usage: []string{
			T(`CF_NAME curl PATH [-iv] [-X METHOD] [-H HEADER] [-d DATA] [--output FILE]

   By default 'CF_NAME curl' will perform a GET to the specified PATH. If data
   is provided via -d, a POST will be performed instead, and the Content-Type
   will be set to application/json. You may override headers with -H and the
   request method with -X.

   For API documentation, please visit http://apidocs.cloudfoundry.org.`),
		},
		Examples: []string{
			`CF_NAME curl "/v2/apps" -X GET -H "Content-Type: application/x-www-form-urlencoded" -d 'q=name:myapp'`,
			`CF_NAME curl "/v2/apps" -d @/path/to/file`,
		},
		Flags: fs,
	}
}

func (cmd *Curl) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. An argument is missing or not correctly enclosed.\n\n") + commandregistry.Commands.CommandUsage("curl"))
	}

	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *Curl) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.curlRepo = deps.RepoLocator.GetCurlRepository()
	return cmd
}

func (cmd *Curl) Execute(c flags.FlagContext) error {
	path := c.Args()[0]
	headers := c.StringSlice("H")

	var method string
	var body string

	if c.IsSet("d") {
		method = "POST"

		jsonBytes, err := util.GetContentsFromOptionalFlagValue(c.String("d"))
		if err != nil {
			return err
		}
		body = string(jsonBytes)
	}

	if c.IsSet("X") {
		method = c.String("X")
	}

	reqHeader := strings.Join(headers, "\n")

	responseHeader, responseBody, apiErr := cmd.curlRepo.Request(method, path, reqHeader, body)
	if apiErr != nil {
		return errors.New(T("Error creating request:\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
	}

	if trace.LoggingToStdout {
		return nil
	}

	if c.Bool("i") {
		cmd.ui.Say(responseHeader)
	}

	if c.String("output") != "" {
		err := cmd.writeToFile(responseBody, c.String("output"))
		if err != nil {
			return errors.New(T("Error creating request:\n{{.Err}}", map[string]interface{}{"Err": err}))
		}
	} else {
		if strings.Contains(responseHeader, "application/json") {
			buffer := bytes.Buffer{}
			err := json.Indent(&buffer, []byte(responseBody), "", "   ")
			if err == nil {
				responseBody = buffer.String()
			}
		}

		cmd.ui.Say(responseBody)
	}
	return nil
}

func (cmd Curl) writeToFile(responseBody, filePath string) (err error) {
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
	}

	if err != nil {
		return
	}

	return ioutil.WriteFile(filePath, []byte(responseBody), 0644)
}
