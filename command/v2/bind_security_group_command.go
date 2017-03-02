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

//go:generate counterfeiter . BindSecurityGroupActor

type BindSecurityGroupActor interface {
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
	GetOrganizationSpaces(orgGUID string) ([]v2action.Space, v2action.Warnings, error)
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
	GetSecurityGroupByName(securityGroupName string) (v2action.SecurityGroup, v2action.Warnings, error)
	BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string) (v2action.Warnings, error)
}

type BindSecurityGroupCommand struct {
	RequiredArgs    flag.BindSecurityGroupArgs `positional-args:"yes"`
	usage           interface{}                `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [SPACE]\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}                `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       BindSecurityGroupActor
}

func (cmd *BindSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, nil)

	return nil
}

func (cmd BindSecurityGroupCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}

	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.SharedActor.CheckTarget(cmd.Config, false, false)
	if err != nil {
		return shared.HandleError(err)
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	securityGroup, warnings, err := cmd.Actor.GetSecurityGroupByName(cmd.RequiredArgs.SecurityGroupName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.OrganizationName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	if cmd.RequiredArgs.SpaceName != "" {
		space, warnings, err := cmd.Actor.GetSpaceByOrganizationAndName(org.GUID, cmd.RequiredArgs.SpaceName)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return shared.HandleError(err)
		}

		warnings, _ = cmd.Actor.BindSecurityGroupToSpace(securityGroup.GUID, space.GUID)
		cmd.UI.DisplayWarnings(warnings)
	}

	// security group,
	// get security group
	// get org
	// get space OR get spaces in org
	// for each space, bind security group
	return nil
}
