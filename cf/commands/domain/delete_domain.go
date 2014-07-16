package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteDomain struct {
	ui         terminal.UI
	config     configuration.Reader
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewDeleteDomain(ui terminal.UI, config configuration.Reader, repo api.DomainRepository) (cmd *DeleteDomain) {
	cmd = new(DeleteDomain)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = repo
	return
}

func (cmd *DeleteDomain) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-domain",
		Description: T("Delete a domain"),
		Usage:       T("CF_NAME delete-domain DOMAIN [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd *DeleteDomain) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	loginReq := requirementsFactory.NewLoginRequirement()
	cmd.orgReq = requirementsFactory.NewTargetedOrgRequirement()

	reqs = []requirements.Requirement{
		loginReq,
		cmd.orgReq,
	}

	return
}

func (cmd *DeleteDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	domain, apiErr := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganizationFields().Guid)

	switch apiErr.(type) {
	case nil: //do nothing
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(apiErr.Error())
		return
	default:
		cmd.ui.Failed(T("Error finding domain {{.DomainName}}\n{{.ApiErr}}",
			map[string]interface{}{"DomainName": domainName, "ApiErr": apiErr.Error()}))
		return
	}

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("domain"), domainName) {
			return
		}
	}

	cmd.ui.Say(T("Deleting domain {{.DomainName}} as {{.ApiErr}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"ApiErr":     terminal.EntityNameColor(cmd.config.Username())}))

	apiErr = cmd.domainRepo.Delete(domain.Guid)
	if apiErr != nil {
		cmd.ui.Failed(T("Error deleting domain {{.DomainName}}\n{{.ApiErr}}",
			map[string]interface{}{"DomainName": domainName, "ApiErr": apiErr.Error()}))
		return
	}

	cmd.ui.Ok()
}
