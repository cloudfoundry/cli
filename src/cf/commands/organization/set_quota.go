package organization

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetQuota struct {
	ui        terminal.UI
	config    *configuration.Configuration
	quotaRepo api.QuotaRepository
	orgReq    requirements.OrganizationRequirement
}

func NewSetQuota(ui terminal.UI, config *configuration.Configuration, quotaRepo api.QuotaRepository) (cmd *SetQuota) {
	cmd = new(SetQuota)
	cmd.ui = ui
	cmd.config = config
	cmd.quotaRepo = quotaRepo
	return
}

func (cmd *SetQuota) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-quota")
		return
	}

	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *SetQuota) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()
	quotaName := c.Args()[1]
	quota, apiResponse := cmd.quotaRepo.FindByName(quotaName)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Say("Setting quota %s to org %s as %s...",
		terminal.EntityNameColor(quota.Name),
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse = cmd.quotaRepo.Update(org, quota)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
