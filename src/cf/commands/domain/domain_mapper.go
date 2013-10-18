package domain

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DomainMapper struct {
	ui         terminal.UI
	domainRepo api.DomainRepository
	spaceReq   requirements.SpaceRequirement
	orgReq     requirements.TargetedOrgRequirement
	bind       bool
}

func NewDomainMapper(ui terminal.UI, domainRepo api.DomainRepository, bind bool) (cmd *DomainMapper) {
	cmd = &DomainMapper{
		ui:         ui,
		domainRepo: domainRepo,
		bind:       bind,
	}
	return
}

func (cmd *DomainMapper) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		if cmd.bind {
			cmd.ui.FailWithUsage(c, "map-domain")
		} else {
			cmd.ui.FailWithUsage(c, "unmap-domain")
		}
		return
	}

	spaceName := c.Args()[0]
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

func (cmd *DomainMapper) Run(c *cli.Context) {
	var (
		apiResponse net.ApiResponse
		domain      cf.Domain
	)

	domainName := c.Args()[1]
	space := cmd.spaceReq.GetSpace()
	org := cmd.orgReq.GetOrganization()

	if cmd.bind {
		cmd.ui.Say("Mapping domain %s to space %s...", domainName, space.Name)
	} else {
		cmd.ui.Say("Unmapping domain %s from space %s...", domainName, space.Name)
	}

	domain, apiResponse = cmd.domainRepo.FindByNameInOrg(domainName, org)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error finding domain %s\n%s", domainName, apiResponse.Message)
		return
	}

	if cmd.bind {
		apiResponse = cmd.domainRepo.Map(domain, space)
	} else {
		apiResponse = cmd.domainRepo.Unmap(domain, space)
	}

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	return
}
