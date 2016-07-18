package organization

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RenameOrg struct {
	ui      terminal.UI
	config  coreconfig.ReadWriter
	orgRepo organizations.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&RenameOrg{})
}

func (cmd *RenameOrg) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "rename-org",
		Description: T("Rename an org"),
		Usage: []string{
			T("CF_NAME rename-org ORG NEW_ORG"),
		},
	}
}

func (cmd *RenameOrg) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires old org name, new org name as arguments\n\n") + commandregistry.Commands.CommandUsage("rename-org"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs
}

func (cmd *RenameOrg) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	return cmd
}

func (cmd *RenameOrg) Execute(c flags.FlagContext) error {
	org := cmd.orgReq.GetOrganization()
	newName := c.Args()[1]

	cmd.ui.Say(T("Renaming org {{.OrgName}} to {{.NewName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(org.Name),
			"NewName":  terminal.EntityNameColor(newName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.orgRepo.Rename(org.GUID, newName)
	if err != nil {
		return err
	}
	cmd.ui.Ok()

	if org.GUID == cmd.config.OrganizationFields().GUID {
		org.Name = newName
		cmd.config.SetOrganizationFields(org.OrganizationFields)
	}
	return nil
}
