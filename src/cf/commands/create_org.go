package commands

import (
	/*	"cf"*/
	/*	"cf/api"*/
	"cf/requirements"
	term "cf/terminal"
	/*	"errors"
		"fmt"*/
	"github.com/codegangsta/cli"
	/*	"strings"*/
)

type CreateOrganization struct {
	ui term.UI
}

func NewCreateOrganization(ui term.UI) (cmd CreateOrganization) {
	cmd.ui = ui
	return
}

func (cmd CreateOrganization) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd CreateOrganization) Run(c *cli.Context) {
	name := c.String("name")
	cmd.createOrganization(name)
	return
}

func (cmd CreateOrganization) createOrganization(name string) {
	cmd.ui.Say("Creating organization %s", term.Cyan(name))
	cmd.ui.Ok()
	return
}
