package organization

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateOrg struct {
	ui      terminal.UI
	config  configuration.Reader
	orgRepo api.OrganizationRepository
}

func NewCreateOrg(ui terminal.UI, config configuration.Reader, orgRepo api.OrganizationRepository) (cmd CreateOrg) {
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	return
}

func (cmd CreateOrg) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-org")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd CreateOrg) Run(c *cli.Context) {
	name := c.Args()[0]

	cmd.ui.Say("Creating org %s as %s...",
		terminal.EntityNameColor(name),
		terminal.EntityNameColor(cmd.config.Username()),
	)
	err := cmd.orgRepo.Create(name)
	if err != nil {
		if apiErr, ok := err.(errors.HttpError); ok && apiErr.ErrorCode() == errors.ORG_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn("Org %s already exists", name)
			return
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("\nTIP: Use '%s' to target new org", terminal.CommandColor(cf.Name()+" target -o "+name))
}
