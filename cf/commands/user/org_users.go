package user

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

var orgRoles = []string{models.ORG_MANAGER, models.BILLING_MANAGER, models.ORG_AUDITOR}

var orgRoleToDisplayName = map[string]string{
	models.ORG_USER:        "USERS",
	models.ORG_MANAGER:     "ORG MANAGER",
	models.BILLING_MANAGER: "BILLING MANAGER",
	models.ORG_AUDITOR:     "ORG AUDITOR",
}

type OrgUsers struct {
	ui       terminal.UI
	config   configuration.Reader
	orgReq   requirements.OrganizationRequirement
	userRepo api.UserRepository
}

func NewOrgUsers(ui terminal.UI, config configuration.Reader, userRepo api.UserRepository) (cmd *OrgUsers) {
	cmd = new(OrgUsers)
	cmd.ui = ui
	cmd.config = config
	cmd.userRepo = userRepo
	return
}

func (command *OrgUsers) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "org-users",
		Description: "Show org users by role",
		Usage:       "CF_NAME org-users ORG",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "a", Usage: "List all users in the org"},
		},
	}
}

func (cmd *OrgUsers) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "org-users")
		return
	}

	orgName := c.Args()[0]
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(orgName)
	reqs = append(reqs, requirementsFactory.NewLoginRequirement(), cmd.orgReq)

	return
}

func (cmd *OrgUsers) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()
	all := c.Bool("a")

	cmd.ui.Say("Getting users in org %s as %s...",
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	roles := orgRoles
	if all {
		roles = []string{models.ORG_USER}
	}

	for _, role := range roles {
		displayName := orgRoleToDisplayName[role]

		users, apiErr := cmd.userRepo.ListUsersInOrgForRole(org.Guid, role)

		cmd.ui.Say("")
		cmd.ui.Say("%s", terminal.HeaderColor(displayName))

		for _, user := range users {
			cmd.ui.Say("  %s", user.Username)
		}

		if apiErr != nil {
			cmd.ui.Failed("Failed fetching org-users for role %s.\n%s", apiErr.Error(), displayName)
			return
		}
	}
}
