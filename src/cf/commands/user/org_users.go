package user

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

var orgRoles = []string{cf.ORG_MANAGER, cf.BILLING_MANAGER, cf.ORG_AUDITOR}

var orgRoleToDisplayName = map[string]string{
	cf.ORG_USER:        "USERS",
	cf.ORG_MANAGER:     "ORG MANAGER",
	cf.BILLING_MANAGER: "BILLING MANAGER",
	cf.ORG_AUDITOR:     "ORG AUDITOR",
}

type OrgUsers struct {
	ui       terminal.UI
	config   *configuration.Configuration
	orgReq   requirements.OrganizationRequirement
	userRepo api.UserRepository
}

func NewOrgUsers(ui terminal.UI, config *configuration.Configuration, userRepo api.UserRepository) (cmd *OrgUsers) {
	cmd = new(OrgUsers)
	cmd.ui = ui
	cmd.config = config
	cmd.userRepo = userRepo
	return
}

func (cmd *OrgUsers) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect usage")
		cmd.ui.FailWithUsage(c, "org-users")
		return
	}

	orgName := c.Args()[0]
	cmd.orgReq = reqFactory.NewOrganizationRequirement(orgName)
	reqs = append(reqs, reqFactory.NewLoginRequirement(), cmd.orgReq)

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
		roles = []string{cf.ORG_USER}
	}

	for _, role := range roles {
		stopChan := make(chan bool)
		defer close(stopChan)

		displayName := orgRoleToDisplayName[role]

		usersChan, statusChan := cmd.userRepo.ListUsersInOrgForRole(org.Guid, role, stopChan)

		cmd.ui.Say("")
		cmd.ui.Say("%s", terminal.HeaderColor(displayName))

		for users := range usersChan {
			for _, user := range users {
				cmd.ui.Say("  %s", user.Username)
			}
		}

		apiStatus := <-statusChan
		if apiStatus.IsNotSuccessful() {
			cmd.ui.Failed("Failed fetching org-users for role %s.\n%s", apiStatus.Message, displayName)
			return
		}
	}
}
