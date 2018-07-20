package v2

import (
	"os"
	"strconv"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . CreateOrgActor

type CreateOrgActor interface {
	CreateOrganization(orgName string) (v2action.Organization, v2action.Warnings, error)
	GrantOrgManagerByUsername(guid string, username string) (v2action.Warnings, error)
}

type CreateOrgCommand struct {
	RequiredArgs    flag.Organization `positional-args:"yes"`
	Quota           string            `short:"q" description:"Quota to assign to the newly created org (excluding this option results in assignment of default quota)"`
	usage           interface{}       `usage:"CF_NAME create-org ORG"`
	relatedCommands interface{}       `related_commands:"create-space, orgs, quotas, set-org-role"`

	UI          command.UI
	Config      command.Config
	Actor       CreateOrgActor
	SharedActor command.SharedActor
}

func (cmd *CreateOrgCommand) Setup(config command.Config, ui command.UI) error {
	if !config.Experimental() {
		return nil
	}
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd CreateOrgCommand) Execute(args []string) error {
	experimental, err := strconv.ParseBool(os.Getenv("CF_CLI_EXPERIMENTAL"))
	if !experimental || err != nil {
		return translatableerror.UnrefactoredCommandError{}
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	orgName := cmd.RequiredArgs.Organization

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  orgName,
			"Username": user.Name,
		})

	org, warnings, err := cmd.Actor.CreateOrganization(orgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayTextWithFlavor("Assigning role OrgManager to user {{.Username}} in org {{.OrgName}}...",
		map[string]interface{}{
			"OrgName":  orgName,
			"Username": user.Name,
		})
	warnings, _ = cmd.Actor.GrantOrgManagerByUsername(org.GUID, user.Name)
	cmd.UI.DisplayWarnings(warnings)

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayText(`TIP: Use 'cf target -o "{{.OrgName}}"' to target new org`,
		map[string]interface{}{
			"OrgName": orgName,
		})
	return nil
}
