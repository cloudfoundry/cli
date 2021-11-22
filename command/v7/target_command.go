package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
)

type TargetCommand struct {
	BaseCommand

	Organization    string      `short:"o" description:"Organization"`
	Space           string      `short:"s" description:"Space"`
	usage           interface{} `usage:"CF_NAME target [-o ORG] [-s SPACE]"`
	relatedCommands interface{} `related_commands:"create-org, create-space, login, orgs, spaces"`
}

func (cmd *TargetCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		cmd.clearTargets()
		return err
	}

	switch {
	case cmd.Organization != "" && cmd.Space != "":
		err = cmd.setOrgAndSpace()
		if err != nil {
			cmd.clearTargets()
			return err
		}
	case cmd.Organization != "":
		err = cmd.setOrg()
		if err != nil {
			cmd.clearTargets()
			return err
		}
		err = cmd.autoTargetSpace(cmd.Config.TargetedOrganization().GUID)
		if err != nil {
			cmd.clearTargets()
			return err
		}
	case cmd.Space != "":
		err = cmd.setSpace()
		if err != nil {
			cmd.clearTargets()
			return err
		}
	}

	cmd.displayTargetTable(user)

	if !cmd.Config.HasTargetedOrganization() {
		cmd.UI.DisplayText("No org or space targeted, use '{{.CFTargetCommand}}'",
			map[string]interface{}{
				"CFTargetCommand": fmt.Sprintf("%s target -o ORG -s SPACE", cmd.Config.BinaryName()),
			})
		return nil
	}

	if !cmd.Config.HasTargetedSpace() {
		cmd.UI.DisplayText("No space targeted, use '{{.CFTargetCommand}}'",
			map[string]interface{}{
				"CFTargetCommand": fmt.Sprintf("%s target -s SPACE", cmd.Config.BinaryName()),
			})
	}

	return nil
}

func (cmd TargetCommand) clearTargets() {
	if cmd.Organization != "" {
		cmd.Config.UnsetOrganizationAndSpaceInformation()
	} else if cmd.Space != "" {
		cmd.Config.UnsetSpaceInformation()
	}
}

// setOrgAndSpace sets organization and space
func (cmd *TargetCommand) setOrgAndSpace() error {
	err := cmd.setOrg()
	if err != nil {
		return err
	}

	err = cmd.setSpace()
	if err != nil {
		return err
	}

	return nil
}

// setOrg sets organization
func (cmd *TargetCommand) setOrg() error {
	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.Config.SetOrganizationInformation(org.GUID, cmd.Organization)
	cmd.Config.UnsetSpaceInformation()

	return nil
}

// autoTargetSpace targets the space if there is only one space in the org
// and no space arg was provided.
func (cmd *TargetCommand) autoTargetSpace(orgGUID string) error {
	spaces, warnings, err := cmd.Actor.GetOrganizationSpaces(orgGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(spaces) == 1 {
		space := spaces[0]
		cmd.Config.V7SetSpaceInformation(space.GUID, space.Name)
	}

	return nil
}

// setSpace sets space
func (cmd *TargetCommand) setSpace() error {
	if !cmd.Config.HasTargetedOrganization() {
		return translatableerror.NoOrganizationTargetedError{BinaryName: cmd.Config.BinaryName()}
	}

	space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.Space, cmd.Config.TargetedOrganization().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.Config.V7SetSpaceInformation(space.GUID, space.Name)

	return nil
}

// displayTargetTable neatly displays target information.
func (cmd *TargetCommand) displayTargetTable(user configv3.User) {
	table := [][]string{
		{cmd.UI.TranslateText("API endpoint:"), cmd.Config.Target()},
		{cmd.UI.TranslateText("API version:"), cmd.Config.APIVersion()},
		{cmd.UI.TranslateText("user:"), user.Name},
	}

	if cmd.Config.HasTargetedOrganization() {
		table = append(table, []string{
			cmd.UI.TranslateText("org:"), cmd.Config.TargetedOrganization().Name,
		})
	}

	if cmd.Config.HasTargetedSpace() {
		table = append(table, []string{
			cmd.UI.TranslateText("space:"), cmd.Config.TargetedSpace().Name,
		})
	}
	cmd.UI.DisplayKeyValueTable("", table, 3)
}
