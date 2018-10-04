package v6

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . CreateSpaceActor

type CreateSpaceActor interface {
	CreateSpace(spaceName, orgName, quotaName string) (v2action.Space, v2action.Warnings, error)
	GrantSpaceManagerByUsername(orgGUID string, spaceGUID string, username string) (v2action.Warnings, error)
	GrantSpaceDeveloperByUsername(spaceGUID string, username string) (v2action.Warnings, error)
}

type CreateSpaceCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	Organization    string      `short:"o" description:"Organization"`
	Quota           string      `short:"q" description:"Quota to assign to the newly created space"`
	usage           interface{} `usage:"CF_NAME create-space SPACE [-o ORG] [-q SPACE_QUOTA]"`
	relatedCommands interface{} `related_commands:"set-space-isolation-segment, space-quotas, spaces, target"`

	UI          command.UI
	Config      command.Config
	Actor       CreateSpaceActor
	SharedActor command.SharedActor
}

func (cmd *CreateSpaceCommand) Setup(config command.Config, ui command.UI) error {
	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.SharedActor = sharedaction.NewActor(config)
	cmd.Config = config
	cmd.UI = ui
	return nil
}

func (cmd CreateSpaceCommand) Execute(args []string) error {
	if !cmd.Config.Experimental() {
		return translatableerror.UnrefactoredCommandError{}
	}

	spaceName := cmd.RequiredArgs.Space
	userName, err := cmd.SharedActor.RequireCurrentUser()
	if err != nil {
		return err
	}

	var orgName string
	if cmd.Organization == "" {
		orgName, err = cmd.SharedActor.RequireTargetedOrg()
		if err != nil {
			return err
		}
	} else {
		orgName = cmd.Organization
	}

	cmd.UI.DisplayTextWithFlavor("Creating space {{.Space}} in org {{.Org}} as {{.User}}...", map[string]interface{}{
		"Space": spaceName,
		"Org":   orgName,
		"User":  userName,
	})

	space, warnings, err := cmd.Actor.CreateSpace(spaceName, orgName, cmd.Quota)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		if _, ok := err.(actionerror.SpaceNameTakenError); ok {
			cmd.UI.DisplayOK()

			cmd.UI.DisplayWarning("Space {{.SpaceName}} already exists", map[string]interface{}{
				"SpaceName": spaceName,
			})
			return nil
		} else {
			return err
		}
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayTextWithFlavor("Assigning role SpaceManager to user {{.User}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"Space": spaceName,
		"Org":   orgName,
		"User":  userName,
	})

	warnings, err = cmd.Actor.GrantSpaceManagerByUsername(space.OrganizationGUID, space.GUID, userName)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Assigning role SpaceDeveloper to user {{.User}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"Space": spaceName,
		"Org":   orgName,
		"User":  userName,
	})

	warnings, err = cmd.Actor.GrantSpaceDeveloperByUsername(space.GUID, userName)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayText(`TIP: Use 'cf target -o "{{.Org}}" -s "{{.Space}}"' to target new space`, map[string]interface{}{
		"Org":   orgName,
		"Space": spaceName,
	})

	return nil
}
