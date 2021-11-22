package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
)

type DeleteSpaceCommand struct {
	BaseCommand

	RequiredArgs flag.Space  `positional-args:"yes"`
	Force        bool        `short:"f" description:"Force deletion without confirmation"`
	Org          string      `short:"o" description:"Delete space within specified org"`
	usage        interface{} `usage:"CF_NAME delete-space SPACE [-o ORG] [-f]"`
}

func (cmd DeleteSpaceCommand) Execute(args []string) error {
	var (
		err     error
		orgName string
	)

	if cmd.Org == "" {
		err = cmd.SharedActor.CheckTarget(true, false)
		orgName = cmd.Config.TargetedOrganization().Name
	} else {
		err = cmd.SharedActor.CheckTarget(false, false)
		orgName = cmd.Org
	}

	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		cmd.UI.DisplayText("This action impacts all resources scoped to this space, including apps, service instances, and space-scoped service brokers.")
		const promptMessage = "Really delete the space {{.SpaceName}}?"
		deleteSpace, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{"SpaceName": cmd.RequiredArgs.Space})

		if promptErr != nil {
			return promptErr
		}

		if !deleteSpace {
			cmd.UI.DisplayText("'{{.TargetSpace}}' has not been deleted.",
				map[string]interface{}{
					"TargetSpace": cmd.RequiredArgs.Space,
				})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting space {{.TargetSpace}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetSpace": cmd.RequiredArgs.Space,
			"TargetOrg":   orgName,
			"CurrentUser": user.Name,
		})

	warnings, err := cmd.Actor.DeleteSpaceByNameAndOrganizationName(cmd.RequiredArgs.Space, orgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.SpaceNotFoundError:
			cmd.UI.DisplayWarning("Space '{{.SpaceName}}' does not exist.", map[string]interface{}{
				"SpaceName": cmd.RequiredArgs.Space,
			})
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	if cmd.Config.TargetedOrganization().Name == orgName &&
		cmd.Config.TargetedSpace().Name == cmd.RequiredArgs.Space {
		cmd.Config.UnsetSpaceInformation()
		cmd.UI.DisplayText("TIP: No space targeted, use '{{.CfTargetCommand}}' to target a space.",
			map[string]interface{}{"CfTargetCommand": cmd.Config.BinaryName() + " target -s"})
	}

	return nil
}
