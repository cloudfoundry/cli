package organization

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetQuota struct {
	ui      terminal.UI
	orgRepo api.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func NewSetQuota(ui terminal.UI, orgRepo api.OrganizationRepository) (cmd *SetQuota) {
	cmd = new(SetQuota)
	cmd.ui = ui
	cmd.orgRepo = orgRepo
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
	quota, apiResponse := cmd.orgRepo.FindQuotaByName(quotaName)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Say("Setting quota %s to org %s...",
		terminal.EntityNameColor(quota.Name),
		terminal.EntityNameColor(org.Name))

	apiResponse = cmd.orgRepo.UpdateQuota(org, quota)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
