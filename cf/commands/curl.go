package commands

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
)

type Curl struct {
	ui       terminal.UI
	config   core_config.Reader
	curlRepo api.CurlRepository
}

func init() {
	command_registry.Register(&Curl{})
}

func (cmd *Curl) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &cliFlags.BoolFlag{ShortName: "i", Usage: T("Include response headers in the output")}
	fs["v"] = &cliFlags.BoolFlag{ShortName: "v", Usage: T("Enable CF_TRACE output for all requests and responses")}
	fs["X"] = &cliFlags.StringFlag{ShortName: "X", Value: "GET", Usage: T("HTTP method (GET,POST,PUT,DELETE,etc)")}
	fs["H"] = &cliFlags.StringSliceFlag{ShortName: "H", Usage: T("Custom headers to include in the request, flag can be specified multiple times")}
	fs["d"] = &cliFlags.StringFlag{ShortName: "d", Usage: T("HTTP data to include in the request body")}
	fs["output"] = &cliFlags.StringFlag{Name: "output", Usage: T("Write curl body to FILE instead of stdout")}

	return command_registry.CommandMetadata{
		Name:        "curl",
		Description: T("Executes a raw request, content-type set to application/json by default"),
		Usage:       T("CF_NAME curl PATH [-iv] [-X METHOD] [-H HEADER] [-d DATA] [--output FILE]") + "\n   " + T("For API documentation, please visit http://apidocs.cloudfoundry.org"),
		Flags:       fs,
	}
}

func (cmd *Curl) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. An argument is missing or not correctly enclosed.\n\n") + command_registry.Commands.CommandUsage("curl"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *Curl) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.curlRepo = deps.RepoLocator.GetCurlRepository()
	return cmd
}

func (cmd *Curl) Execute(c flags.FlagContext) {
	path := c.Args()[0]
	method := c.String("X")
	headers := c.StringSlice("H")
	body := c.String("d")
	verbose := c.Bool("v")

	reqHeader := strings.Join(headers, "\n")

	if verbose {
		trace.EnableTrace()
	}

	responseHeader, responseBody, apiErr := cmd.curlRepo.Request(method, path, reqHeader, body)
	if apiErr != nil {
		cmd.ui.Failed(T("Error creating request:\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
	}

	if verbose {
		return
	}

	if c.Bool("i") {
		cmd.ui.Say(responseHeader)
	}

	if c.String("output") != "" {
		err := cmd.writeToFile(responseBody, c.String("output"))
		if err != nil {
			cmd.ui.Failed(T("Error creating request:\n{{.Err}}", map[string]interface{}{"Err": err}))
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
	return
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
