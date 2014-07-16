package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateDomain struct {
	ui         terminal.UI
	config     configuration.Reader
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func NewCreateDomain(ui terminal.UI, config configuration.Reader, domainRepo api.DomainRepository) (cmd *CreateDomain) {
	cmd = new(CreateDomain)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = domainRepo
	return
}

func (cmd *CreateDomain) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-domain",
		Description: T("Create a domain in an org for later use"),
		Usage:       T("CF_NAME create-domain ORG DOMAIN"),
	}
}

func (cmd *CreateDomain) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *CreateDomain) Run(c *cli.Context) {
	domainName := c.Args()[1]
	owningOrg := cmd.orgReq.GetOrganization()

	cmd.ui.Say(T("Creating domain {{.DomainName}} for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"OrgName":    terminal.EntityNameColor(owningOrg.Name),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	_, apiErr := cmd.domainRepo.Create(domainName, owningOrg.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
