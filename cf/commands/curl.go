package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/codegangsta/cli"
	"strings"
)

type Curl struct {
	ui       terminal.UI
	config   configuration.Reader
	curlRepo api.CurlRepository
}

func NewCurl(ui terminal.UI, config configuration.Reader, curlRepo api.CurlRepository) (cmd *Curl) {
	cmd = new(Curl)
	cmd.ui = ui
	cmd.config = config
	cmd.curlRepo = curlRepo
	return
}

func (cmd *Curl) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "curl",
		Description: T("Executes a raw request, content-type set to application/json by default"),
		Usage:       T("CF_NAME curl PATH [-X METHOD] [-H HEADER] [-d DATA] [-i]"),
		Flags: []cli.Flag{
			cli.StringFlag{Name: "X", Value: "GET", Usage: T("HTTP method (GET,POST,PUT,DELETE,etc)")},
			flag_helpers.NewStringSliceFlag("H", T("Custom headers to include in the request, flag can be specified multiple times")),
			flag_helpers.NewStringFlag("d", T("HTTP data to include in the request body")),
			cli.BoolFlag{Name: "i", Usage: T("Include response headers in the output")},
			cli.BoolFlag{Name: "v", Usage: T("Enable CF_TRACE output for all requests and responses")},
		},
	}
}

func (cmd *Curl) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New(T("Incorrect number of arguments"))
		cmd.ui.FailWithUsage(c)
		return
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *Curl) Run(c *cli.Context) {
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
		return
	}

	if verbose {
		return
	}

	if c.Bool("i") {
		cmd.ui.Say(responseHeader)
	}

	if strings.Contains(responseHeader, "application/json") {
		buffer := bytes.Buffer{}
		err := json.Indent(&buffer, []byte(responseBody), "", "   ")
		if err == nil {
			responseBody = buffer.String()
		}
	}

	cmd.ui.Say(responseBody)

	return
}
