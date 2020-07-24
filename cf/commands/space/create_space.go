package space

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spacequotas"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/user"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type CreateSpace struct {
	ui              terminal.UI
	config          coreconfig.Reader
	spaceRepo       spaces.SpaceRepository
	orgRepo         organizations.OrganizationRepository
	userRepo        api.UserRepository
	spaceRoleSetter user.SpaceRoleSetter
	spaceQuotaRepo  spacequotas.SpaceQuotaRepository
}

func init() {
	commandregistry.Register(&CreateSpace{})
}

func (cmd *CreateSpace) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["o"] = &flags.StringFlag{ShortName: "o", Usage: T("Organization")}
	fs["q"] = &flags.StringFlag{ShortName: "q", Usage: T("Quota to assign to the newly created space")}

	return commandregistry.CommandMetadata{
		Name:        "create-space",
		Description: T("Create a space"),
		Usage: []string{
			T("CF_NAME create-space SPACE [-o ORG] [-q SPACE-QUOTA]"),
		},
		Flags: fs,
	}
}

func (cmd *CreateSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("create-space"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.String("o") == "" {
		reqs = append(reqs, requirementsFactory.NewTargetedOrgRequirement())
	}

	return reqs, nil
}

func (cmd *CreateSpace) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()

	//get command from registry for dependency
	commandDep := commandregistry.Commands.FindCommand("set-space-role")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.spaceRoleSetter = commandDep.(user.SpaceRoleSetter)

	return cmd
}

func (cmd *CreateSpace) Execute(c flags.FlagContext) error {
	spaceName := c.Args()[0]
	orgName := c.String("o")
	spaceQuotaName := c.String("q")
	orgGUID := ""
	if orgName == "" {
		orgName = cmd.config.OrganizationFields().Name
		orgGUID = cmd.config.OrganizationFields().GUID
	}

	cmd.ui.Say(T("Creating space {{.SpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"SpaceName":   terminal.EntityNameColor(spaceName),
			"OrgName":     terminal.EntityNameColor(orgName),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	if orgGUID == "" {
		org, err := cmd.orgRepo.FindByName(orgName)
		switch err.(type) {
		case nil:
		case *errors.ModelNotFoundError:
			return errors.New(T("Org {{.OrgName}} does not exist or is not accessible", map[string]interface{}{"OrgName": orgName}))
		default:
			return errors.New(T("Error finding org {{.OrgName}}\n{{.ErrorDescription}}",
				map[string]interface{}{
					"OrgName":          orgName,
					"ErrorDescription": err.Error(),
				}))
		}

		orgGUID = org.GUID
	}

	var spaceQuotaGUID string
	if spaceQuotaName != "" {
		spaceQuota, err := cmd.spaceQuotaRepo.FindByNameAndOrgGUID(spaceQuotaName, orgGUID)
		if err != nil {
			return err
		}
		spaceQuotaGUID = spaceQuota.GUID
	}

	space, err := cmd.spaceRepo.Create(spaceName, orgGUID, spaceQuotaGUID)
	if err != nil {
		if httpErr, ok := err.(errors.HTTPError); ok && httpErr.ErrorCode() == errors.SpaceNameTaken {
			cmd.ui.Ok()
			cmd.ui.Warn(T("Space {{.SpaceName}} already exists", map[string]interface{}{"SpaceName": spaceName}))
			return nil
		}
		return err
	}
	cmd.ui.Ok()

	err = cmd.spaceRoleSetter.SetSpaceRole(space, orgGUID, orgName, models.RoleSpaceManager, cmd.config.UserGUID(), cmd.config.Username())
	if err != nil {
		return err
	}

	err = cmd.spaceRoleSetter.SetSpaceRole(space, orgGUID, orgName, models.RoleSpaceDeveloper, cmd.config.UserGUID(), cmd.config.Username())
	if err != nil {
		return err
	}

	cmd.ui.Say(T("\nTIP: Use '{{.CFTargetCommand}}' to target new space",
		map[string]interface{}{
			"CFTargetCommand": terminal.CommandColor(cf.Name + " target -o \"" + orgName + "\" -s \"" + space.Name + "\""),
		}))
	return nil
}
