package user

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type CFUsers struct {
	ui        terminal.UI
	config    core_config.Reader
	orgRepo   organizations.OrganizationRepository
	spaceRepo spaces.SpaceRepository
}

func ShowUserInfo(ui terminal.UI, config core_config.Reader, orgRepo organizations.OrganizationRepository, spaceRepo spaces.SpaceRepository) (cmd *CFUsers) {
	cmd = new(CFUsers)
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd *CFUsers) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "user-info",
		Description: T("Show user-info with roles"),
		Usage:       T("CF_NAME user-info"),
		Flags:       []cli.Flag{},
	}
}

func (cmd *CFUsers) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 0 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd *CFUsers) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting user information..."))

	table := terminal.NewTable(cmd.ui, []string{T("User"), T("Org"), T("Space"), T("Role")})
	retrivedRoles := []string{} //Variable to store all the roles

	//Fetching the Roles at Organization level

	var orgName = cmd.config.OrganizationFields().Name

	var orgRoleToDisplayName = map[string]string{
		models.ORG_MANAGER:     T("ORG MANAGER"),
		models.BILLING_MANAGER: T("BILLING MANAGER"),
		models.ORG_AUDITOR:     T("ORG AUDITOR"),
	}

	for apiName, displayName := range models.QueryParmToOrgRole {
		org, apiErr := cmd.orgRepo.GetOrgRoleForUser(apiName, orgName)

		if apiErr != nil {
			cmd.ui.Failed(T("Failed fetching org-users for role {{.OrgRoleToDisplayName}}.\n{{.Error}}",
				map[string]interface{}{
					"Error":                apiErr.Error(),
					"OrgRoleToDisplayName": orgRoleToDisplayName[displayName],
				}))
			return
		}

		if org.Name != "" {
			retrivedRoles = append(retrivedRoles, orgRoleToDisplayName[displayName])
		}

	}

	// Fetching space roles
	// If the space is not targeted then space roles will not be fetched
	if cmd.config.SpaceFields().Name != "" {
		var spaceRoleToDisplayName = map[string]string{
			models.SPACE_MANAGER:   T("SPACE MANAGER"),
			models.SPACE_DEVELOPER: T("SPACE DEVELOPER"),
			models.SPACE_AUDITOR:   T("SPACE AUDITOR"),
		}

		for apiName, displayName := range models.QueryParmToSpaceRole {

			apiErr := cmd.spaceRepo.GetSpaceRole(func(space models.Space) bool {
				if space.Name != "" {
					retrivedRoles = append(retrivedRoles, spaceRoleToDisplayName[displayName])
				}
				return true
			}, apiName)
			if apiErr != nil {
				cmd.ui.Failed(T("Failed fetching space-users for role {{.SpaceRoleToDisplayName}}.\n{{.Error}}",
					map[string]interface{}{
						"Error":                  apiErr.Error(),
						"SpaceRoleToDisplayName": spaceRoleToDisplayName[displayName],
					}))
				return
			}
		}
	}
	table.Add(cmd.config.Username(), cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name, strings.Join(retrivedRoles, ", "))
	table.Print()
}
