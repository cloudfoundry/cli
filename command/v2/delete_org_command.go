package v2

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . DeleteOrganizationActor

type DeleteOrganizationActor interface {
	DeleteOrganization(orgName string) (v2action.Warnings, error)
	ClearOrganizationAndSpace(config v2action.Config)
}

type DeleteOrgCommand struct {
	RequiredArgs flag.Organization `positional-args:"yes"`
	Force        bool              `short:"f" description:"Force deletion without confirmation"`
	usage        interface{}       `usage:"CF_NAME delete-org ORG [-f]"`

	Config      command.Config
	UI          command.UI
	SharedActor command.SharedActor
	Actor       DeleteOrganizationActor
}

func (cmd *DeleteOrgCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd *DeleteOrgCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		promptMessage := "Really delete the org {{.OrgName}}, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers?"
		deleteOrg, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{"OrgName": cmd.RequiredArgs.Organization})

		if promptErr != nil {
			return promptErr
		}

		if !deleteOrg {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"OrgName":  cmd.RequiredArgs.Organization,
		"Username": user.Name,
	})

	warnings, err := cmd.Actor.DeleteOrganization(cmd.RequiredArgs.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.OrganizationNotFoundError:
			cmd.UI.DisplayText("Org {{.OrgName}} does not exist.", map[string]interface{}{
				"OrgName": cmd.RequiredArgs.Organization,
			})
		default:
			return err
		}
	}

	if cmd.Config.TargetedOrganization().Name == cmd.RequiredArgs.Organization {
		cmd.Actor.ClearOrganizationAndSpace(cmd.Config)
	}

	cmd.UI.DisplayOK()

	return nil
}
