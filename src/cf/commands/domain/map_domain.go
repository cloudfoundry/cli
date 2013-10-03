package domain

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type MapDomain struct {
	ui         terminal.UI
	domainRepo api.DomainRepository
	spaceReq   requirements.SpaceRequirement
	orgReq     requirements.TargetedOrgRequirement
}

func NewMapDomain(ui terminal.UI, domainRepo api.DomainRepository) (cmd *MapDomain) {
	cmd = &MapDomain{
		ui:         ui,
		domainRepo: domainRepo,
	}
	return
}

func (cmd *MapDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "map-domain")
		return
	}

	spaceName := c.Args()[1]
	cmd.spaceReq = reqFactory.NewSpaceRequirement(spaceName)

	loginReq := reqFactory.NewLoginRequirement()
	cmd.orgReq = reqFactory.NewTargetedOrgRequirement()

	reqs = []requirements.Requirement{
		loginReq,
		cmd.orgReq,
		cmd.spaceReq,
	}

	return
}

func (cmd *MapDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	space := cmd.spaceReq.GetSpace()

	cmd.ui.Say("Mapping domain %s to space %s...", domainName, space.Name)

	org := cmd.orgReq.GetOrganization()
	domain, apiStatus := cmd.domainRepo.FindByNameInOrg(domainName, org)
	if apiStatus.IsError() {
		cmd.ui.Failed("Error finding domain %s\n%s", domainName, apiStatus.Message)
		return
	}
	if apiStatus.IsNotFound() {
		cmd.ui.Failed("Domain %s does not exist", domainName)
		return
	}

	apiStatus = cmd.domainRepo.MapDomain(domain, space)
	if apiStatus.IsError() {
		cmd.ui.Failed("%s", apiStatus.Message)
		return
	}
	cmd.ui.Ok()
	return
}
