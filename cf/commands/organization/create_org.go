package organization

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type CreateOrg struct {
	ui        terminal.UI
	config    core_config.Reader
	orgRepo   organizations.OrganizationRepository
	quotaRepo quotas.QuotaRepository
}

func init() {
	command_registry.Register(&CreateOrg{})
}

func (cmd *CreateOrg) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["q"] = &cliFlags.StringFlag{Name: "q", Usage: T("Quota to assign to the newly created org (excluding this option results in assignment of default quota)")}

	return command_registry.CommandMetadata{
		Name:        "create-org",
		ShortName:   "co",
		Description: T("Create an org"),
		Usage:       T("CF_NAME create-org ORG"),
		Flags:       fs,
	}
}

func (cmd *CreateOrg) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("create-org"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *CreateOrg) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	return cmd
}

func (cmd *CreateOrg) Execute(c flags.FlagContext) {
	name := c.Args()[0]
	cmd.ui.Say(T("Creating org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	org := models.Organization{OrganizationFields: models.OrganizationFields{Name: name}}

	quotaName := c.String("q")
	if quotaName != "" {
		quota, err := cmd.quotaRepo.FindByName(quotaName)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}

		org.QuotaDefinition.Guid = quota.Guid
	}

	err := cmd.orgRepo.Create(org)
	if err != nil {
		if apiErr, ok := err.(errors.HttpError); ok && apiErr.ErrorCode() == errors.ORG_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn(T("Org {{.OrgName}} already exists",
				map[string]interface{}{"OrgName": name}))
			return
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("\nTIP: Use '{{.Command}}' to target new org",
		map[string]interface{}{"Command": terminal.CommandColor(cf.Name() + " target -o " + name)}))
}
