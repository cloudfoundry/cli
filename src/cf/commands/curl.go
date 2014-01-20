package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type Curl struct {
	ui       terminal.UI
	config   *configuration.Configuration
	curlRepo api.CurlRepository
}

func NewCurl(ui terminal.UI, config *configuration.Configuration, curlRepo api.CurlRepository) (cmd *Curl) {
	cmd = new(Curl)
	cmd.ui = ui
	cmd.config = config
	cmd.curlRepo = curlRepo
	return
}

func (cmd *Curl) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect number of arguments")
		cmd.ui.FailWithUsage(c, "curl")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *Curl) Run(c *cli.Context) {
	path := c.Args()[0]
	method := c.String("X")
	headers := c.StringSlice("H")
	body := c.String("d")

	reqHeader := strings.Join(headers, "\n")

	respHeader, respBody, _ := cmd.curlRepo.Request(method, path, reqHeader, body)

	if c.Bool("i") {
		cmd.ui.Say("%s", respHeader)
	}

	cmd.ui.Say("%s", respBody)
	return
}
