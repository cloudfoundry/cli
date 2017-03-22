package v2

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/util/configv3"
)

//go:generate counterfeiter . TargetActor
type TargetActor interface {
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
	GetOrganizationSpaces(orgGUID string) ([]v2action.Space, v2action.Warnings, error)
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
}

type TargetCommand struct {
	Organization    string      `short:"o" description:"Organization"`
	Space           string      `short:"s" description:"Space"`
	usage           interface{} `usage:"CF_NAME target [-o ORG] [-s SPACE]"`
	relatedCommands interface{} `related_commands:"create-org, create-space, login, orgs, spaces"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       TargetActor
}

func (cmd *TargetCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient)

	return nil
}

func (cmd *TargetCommand) Execute(args []string) error {
	cmd.notifyCLIUpdateIfNeeded()

	err := cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		cmd.clearTargets()
		return shared.HandleError(err)
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

func (cmd *TargetCommand) notifyCLIUpdateIfNeeded() {
	err := command.MinimumAPIVersionCheck(cmd.Config.BinaryVersion(), cmd.Config.MinCLIVersion())

	if _, ok := err.(command.MinimumAPIVersionNotMetError); ok {
		cmd.UI.DisplayWarning("Cloud Foundry API version {{.APIVersion}} requires CLI version {{.MinCLIVersion}}. You are currently on version {{.BinaryVersion}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
			map[string]interface{}{
				"APIVersion":    cmd.Config.APIVersion(),
				"MinCLIVersion": cmd.Config.MinCLIVersion(),
				"BinaryVersion": cmd.Config.BinaryVersion(),
			})
	}
}

func (cmd TargetCommand) clearTargets() {
	if cmd.Organization != "" {
		cmd.Config.UnsetOrganizationInformation()
		cmd.Config.UnsetSpaceInformation()
	} else if cmd.Space != "" {
		cmd.Config.UnsetSpaceInformation()
	}
}

// setOrgAndSpace sets organization and space
func (cmd *TargetCommand) setOrgAndSpace() error {
	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	space, warnings, err := cmd.Actor.GetSpaceByOrganizationAndName(org.GUID, cmd.Space)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.Config.SetOrganizationInformation(org.GUID, cmd.Organization)
	cmd.Config.SetSpaceInformation(space.GUID, space.Name, space.AllowSSH)

	return nil
}

// setOrg sets organization
func (cmd *TargetCommand) setOrg() error {
	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.Organization)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
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
		return shared.HandleError(err)
	}

	if len(spaces) == 1 {
		space := spaces[0]
		cmd.Config.SetSpaceInformation(space.GUID, space.Name, space.AllowSSH)
	}

	return nil
}

// setSpace sets space
func (cmd *TargetCommand) setSpace() error {
	if !cmd.Config.HasTargetedOrganization() {
		return shared.NoOrganizationTargetedError{}
	}

	space, warnings, err := cmd.Actor.GetSpaceByOrganizationAndName(cmd.Config.TargetedOrganization().GUID, cmd.Space)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.Config.SetSpaceInformation(space.GUID, space.Name, space.AllowSSH)

	return nil
}

// displayTargetTable neatly displays target information.
func (cmd *TargetCommand) displayTargetTable(user configv3.User) {
	table := [][]string{
		{cmd.UI.TranslateText("api endpoint:"), cmd.Config.Target()},
		{cmd.UI.TranslateText("api version:"), cmd.Config.APIVersion()},
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
	cmd.UI.DisplayTable("", table, 3)
}
