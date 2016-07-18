package organization

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/featureflags"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type CreateOrg struct {
	ui            terminal.UI
	config        coreconfig.Reader
	orgRepo       organizations.OrganizationRepository
	quotaRepo     quotas.QuotaRepository
	orgRoleSetter user.OrgRoleSetter
	flagRepo      featureflags.FeatureFlagRepository
}

func init() {
	commandregistry.Register(&CreateOrg{})
}

func (cmd *CreateOrg) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["q"] = &flags.StringFlag{ShortName: "q", Usage: T("Quota to assign to the newly created org (excluding this option results in assignment of default quota)")}

	return commandregistry.CommandMetadata{
		Name:        "create-org",
		ShortName:   "co",
		Description: T("Create an org"),
		Usage: []string{
			T("CF_NAME create-org ORG"),
		},
		Flags: fs,
	}
}

func (cmd *CreateOrg) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("create-org"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *CreateOrg) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.quotaRepo = deps.RepoLocator.GetQuotaRepository()
	cmd.flagRepo = deps.RepoLocator.GetFeatureFlagRepository()

	//get command from registry for dependency
	commandDep := commandregistry.Commands.FindCommand("set-org-role")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.orgRoleSetter = commandDep.(user.OrgRoleSetter)

	return cmd
}

func (cmd *CreateOrg) Execute(c flags.FlagContext) error {
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
			return err
		}

		org.QuotaDefinition.GUID = quota.GUID
	}

	err := cmd.orgRepo.Create(org)
	if err != nil {
		if apiErr, ok := err.(errors.HTTPError); ok && apiErr.ErrorCode() == errors.OrganizationNameTaken {
			cmd.ui.Ok()
			cmd.ui.Warn(T("Org {{.OrgName}} already exists",
				map[string]interface{}{"OrgName": name}))
			return nil
		}

		return err
	}

	cmd.ui.Ok()

	if cmd.config.IsMinAPIVersion(cf.SetRolesByUsernameMinimumAPIVersion) {
		setRolesByUsernameFlag, err := cmd.flagRepo.FindByName("set_roles_by_username")
		if err != nil {
			cmd.ui.Warn(T("Warning: accessing feature flag 'set_roles_by_username'") + " - " + err.Error() + "\n" + T("Skip assigning org role to user"))
		}

		if setRolesByUsernameFlag.Enabled {
			org, err := cmd.orgRepo.FindByName(name)
			if err != nil {
				return errors.New(T("Error accessing org {{.OrgName}} for GUID': ", map[string]interface{}{"Orgname": name}) + err.Error() + "\n" + T("Skip assigning org role to user"))
			}

			cmd.ui.Say("")
			cmd.ui.Say(T("Assigning role {{.Role}} to user {{.CurrentUser}} in org {{.TargetOrg}} ...",
				map[string]interface{}{
					"Role":        terminal.EntityNameColor("OrgManager"),
					"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
					"TargetOrg":   terminal.EntityNameColor(name),
				}))

			err = cmd.orgRoleSetter.SetOrgRole(org.GUID, models.RoleOrgManager, "", cmd.config.Username())
			if err != nil {
				return errors.New(T("Failed assigning org role to user: ") + err.Error())
			}

			cmd.ui.Ok()
		}
	}

	cmd.ui.Say(T("\nTIP: Use '{{.Command}}' to target new org",
		map[string]interface{}{"Command": terminal.CommandColor(cf.Name + " target -o " + name)}))
	return nil
}
