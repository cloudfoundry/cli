package v2

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/util/configv3"
)

//go:generate counterfeiter . TargetActor
type TargetActor interface {
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
	GetOrganizationSpaces(orgGUID string) ([]v2action.Space, v2action.Warnings, error)
	GetSpaceByName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
}

type TargetCommand struct {
	Organization    string      `short:"o" description:"Organization"`
	Space           string      `short:"s" description:"Space"`
	usage           interface{} `usage:"CF_NAME target [-o ORG] [-s SPACE]"`
	relatedCommands interface{} `related_commands:"create-org, create-space, login, orgs, spaces"`

	Config command.Config
	Actor  TargetActor
	UI     command.UI
}

func (cmd *TargetCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui

	ccClient, uaaClient, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd *TargetCommand) Execute(args []string) error {
	err := command.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.CurrentUserError{
			Message: err.Error(),
		}
	}

	err = cmd.setOrg()
	if err != nil {
		return err
	}

	err = cmd.setSpace()
	if err != nil {
		return err
	}

	return cmd.displayTargetTable(user)
}

// Setting space
func (cmd *TargetCommand) setSpace() error {
	if cmd.Space == "" {
		return nil
	}
	emptyOrg := configv3.Organization{}
	if cmd.Config.TargetedOrganization() == emptyOrg {
		return shared.NoOrgTargetedError{}
	}

	space, warnings, err := cmd.Actor.GetSpaceByName(cmd.Config.TargetedOrganization().GUID, cmd.Space)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return shared.SpaceTargetError{
			Message:   err.Error(),
			SpaceName: cmd.Space,
		}
	}

	cmd.Config.SetSpaceInformation(space.GUID, space.Name, space.AllowSSH)
	return nil
}

// Setting organization
func (cmd *TargetCommand) setOrg() error {
	if cmd.Organization == "" {
		return nil
	}

	var (
		org      v2action.Organization
		warnings v2action.Warnings
	)

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Organization)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return shared.OrgTargetError{
			Message: err.Error(),
		}
	}

	cmd.Config.SetOrganizationInformation(org.GUID, cmd.Organization)

	err = cmd.autoTargetSpace(org.GUID)
	if err != nil {
		return err
	}

	return nil
}

// Auto-target the space if there is only one space in the org
// and no space arg was provided.
func (cmd *TargetCommand) autoTargetSpace(orgGUID string) error {
	// if the user provided a space as an argument, we don't want to autoTarget
	if cmd.Space != "" {
		return nil
	}

	spaces, warnings, err := cmd.Actor.GetOrganizationSpaces(orgGUID)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return shared.GetOrgSpacesError{
			Message: err.Error(),
		}
	}

	if len(spaces) == 1 {
		space := spaces[0]
		cmd.Config.SetSpaceInformation(space.GUID, space.Name, space.AllowSSH)
	}

	return nil
}

func (cmd *TargetCommand) displayTargetTable(user configv3.User) error {
	apiEndpoint := cmd.UI.TranslateText("{{.APIEndpoint}} (API version: {{.APIVersionString}})", map[string]interface{}{
		"APIEndpoint":      cmd.Config.Target(),
		"APIVersionString": cmd.Config.APIVersion(),
	})

	table := [][]string{
		{cmd.UI.TranslateText("API endpoint:"), apiEndpoint},
		{cmd.UI.TranslateText("User:"), user.Name},
	}

	emptyOrg := configv3.Organization{}
	if cmd.Config.TargetedOrganization() == emptyOrg {
		cmd.UI.DisplayTable("", table, 3)
		command := fmt.Sprintf("%s target -o ORG -s SPACE", cmd.Config.BinaryName())

		cmd.UI.DisplayTextWithFlavor("No org or space targeted, use '{{.CFTargetCommand}}'",
			map[string]interface{}{
				"CFTargetCommand": command,
			})
		return nil
	}

	table = append(table, []string{
		cmd.UI.TranslateText("Org:"), cmd.Config.TargetedOrganization().Name,
	})

	emptySpace := configv3.Space{}
	if cmd.Config.TargetedSpace() == emptySpace {
		spaceCommand := fmt.Sprintf("%s target -s SPACE", cmd.Config.BinaryName())

		noSpaceTargeted := cmd.UI.TranslateText("No space targeted, use '{{.CFTargetCommand}}'",
			map[string]interface{}{
				"CFTargetCommand": spaceCommand,
			})

		table = append(table, []string{
			cmd.UI.TranslateText("Space:"), noSpaceTargeted,
		})
	} else {
		table = append(table, []string{
			cmd.UI.TranslateText("Space:"), cmd.Config.TargetedSpace().Name,
		})
	}

	cmd.UI.DisplayTable("", table, 3)

	return nil
}
