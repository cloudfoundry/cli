package organization

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RenameOrg struct {
	ui      terminal.UI
	config  core_config.ReadWriter
	orgRepo organizations.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func init() {
	command_registry.Register(&RenameOrg{})
}

func (cmd *RenameOrg) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "rename-org",
		Description: T("Rename an org"),
		Usage:       T("CF_NAME rename-org ORG NEW_ORG"),
	}
}

func (cmd *RenameOrg) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires old org name, new org name as arguments\n\n") + command_registry.Commands.CommandUsage("rename-org"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *RenameOrg) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	return cmd
}

func (cmd *RenameOrg) Execute(c flags.FlagContext) {
	org := cmd.orgReq.GetOrganization()
	newName := c.Args()[1]

	cmd.ui.Say(T("Renaming org {{.OrgName}} to {{.NewName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(org.Name),
			"NewName":  terminal.EntityNameColor(newName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	apiErr := cmd.orgRepo.Rename(org.Guid, newName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()

	if org.Guid == cmd.config.OrganizationFields().Guid {
		org.Name = newName
		cmd.config.SetOrganizationFields(org.OrganizationFields)
	}
}
