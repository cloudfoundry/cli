package organization

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/quotas"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SetQuota struct {
	ui        terminal.UI
	config    coreconfig.Reader
	quotaRepo quotas.QuotaRepository
	orgReq    requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&SetQuota{})
}

func (cmd *SetQuota) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "set-quota",
		Description: T("Assign a quota to an org"),
		Usage: []string{
			T("CF_NAME set-quota ORG QUOTA\n\n"),
			T("TIP:\n"),
			T("   View allowable quotas with 'CF_NAME quotas'"),
		},
	}
}

func (cmd *SetQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires ORG_NAME, QUOTA as arguments\n\n") + commandregistry.Commands.CommandUsage("set-quota"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs, nil
}

func (cmd *SetQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *SetQuota) Execute(c flags.FlagContext) error {
	org := cmd.orgReq.GetOrganization()
	quotaName := c.Args()[1]
	quota, err := cmd.quotaRepo.FindByName(quotaName)

	if err != nil {
		return err
	}

	cmd.ui.Say(T("Setting quota {{.QuotaName}} to org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"QuotaName": terminal.EntityNameColor(quota.Name),
			"OrgName":   terminal.EntityNameColor(org.Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	err = cmd.quotaRepo.AssignQuotaToOrg(org.GUID, quota.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
