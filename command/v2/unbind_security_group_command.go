package v2

import (
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . UnbindSecurityGroupActor

type UnbindSecurityGroupActor interface {
	UnbindSecurityGroupByNameAndSpace(securityGroupName string, spaceGUID string) (v2action.Warnings, error)
	UnbindSecurityGroupByNameOrganizationNameAndSpaceName(securityGroupName string, orgName string, spaceName string) (v2action.Warnings, error)
}

type UnbindSecurityGroupCommand struct {
	RequiredArgs    flag.UnbindSecurityGroupArgs `positional-args:"yes"`
	usage           interface{}                  `usage:"CF_NAME unbind-security-group SECURITY_GROUP ORG SPACE\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}                  `related_commands:"apps, restart, security-groups"`

	UI          command.UI
	Config      command.Config
	Actor       UnbindSecurityGroupActor
	SharedActor command.SharedActor
}

func (cmd *UnbindSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, nil)

	return nil
}

func (cmd UnbindSecurityGroupCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	var (
		warnings v2action.Warnings
	)

	switch {
	case cmd.RequiredArgs.OrganizationName == "" && cmd.RequiredArgs.SpaceName == "":
		err = cmd.SharedActor.CheckTarget(cmd.Config, true, true)
		if err != nil {
			return shared.HandleError(err)
		}

		space := cmd.Config.TargetedSpace()
		cmd.UI.DisplayTextWithFlavor("Unbinding security group {{.SecurityGroupName}} from {{.OrgName}}/{{.SpaceName}} as {{.Username}}", map[string]interface{}{
			"SecurityGroupName": cmd.RequiredArgs.SecurityGroupName,
			"OrgName":           cmd.Config.TargetedOrganization().Name,
			"SpaceName":         space.Name,
			"Username":          user.Name,
		})
		warnings, err = cmd.Actor.UnbindSecurityGroupByNameAndSpace(cmd.RequiredArgs.SecurityGroupName, space.GUID)

	case cmd.RequiredArgs.OrganizationName != "" && cmd.RequiredArgs.SpaceName != "":
		err = cmd.SharedActor.CheckTarget(cmd.Config, false, false)
		if err != nil {
			return shared.HandleError(err)
		}

		cmd.UI.DisplayTextWithFlavor("Unbinding security group {{.SecurityGroupName}} from {{.OrgName}}/{{.SpaceName}} as {{.Username}}", map[string]interface{}{
			"SecurityGroupName": cmd.RequiredArgs.SecurityGroupName,
			"OrgName":           cmd.RequiredArgs.OrganizationName,
			"SpaceName":         cmd.RequiredArgs.SpaceName,
			"Username":          user.Name,
		})
		warnings, err = cmd.Actor.UnbindSecurityGroupByNameOrganizationNameAndSpaceName(cmd.RequiredArgs.SecurityGroupName, cmd.RequiredArgs.OrganizationName, cmd.RequiredArgs.SpaceName)

	default:
		return command.ThreeRequiredArgumentsError{
			ArgumentName1: "SECURITY_GROUP",
			ArgumentName2: "ORG",
			ArgumentName3: "SPACE"}
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("TIP: Changes will not apply to existing running applications until they are restarted.")

	return nil
}
