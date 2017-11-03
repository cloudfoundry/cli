package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . BindSecurityGroupActor

type BindSecurityGroupActor interface {
	BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string, lifecycle ccv2.SecurityGroupLifecycle) (v2action.Warnings, error)
	CloudControllerAPIVersion() string
	GetOrganizationByName(orgName string) (v2action.Organization, v2action.Warnings, error)
	GetOrganizationSpaces(orgGUID string) ([]v2action.Space, v2action.Warnings, error)
	GetSecurityGroupByName(securityGroupName string) (v2action.SecurityGroup, v2action.Warnings, error)
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
}

type BindSecurityGroupCommand struct {
	RequiredArgs    flag.BindSecurityGroupArgs  `positional-args:"yes"`
	Lifecycle       flag.SecurityGroupLifecycle `long:"lifecycle" choice:"running" choice:"staging" default:"running" description:"Lifecycle phase the group applies to"`
	usage           interface{}                 `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [SPACE] [--lifecycle (running | staging)]\n\nTIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}                 `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       BindSecurityGroupActor
}

func (cmd *BindSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd BindSecurityGroupCommand) Execute(args []string) error {
	var err error
	if ccv2.SecurityGroupLifecycle(cmd.Lifecycle) == ccv2.SecurityGroupLifecycleStaging {
		err = command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionLifecyleStagingV2)
		if err != nil {
			switch e := err.(type) {
			case translatableerror.MinimumAPIVersionNotMetError:
				return translatableerror.LifecycleMinimumAPIVersionNotMetError{
					CurrentVersion: e.CurrentVersion,
					MinimumVersion: e.MinimumVersion,
				}
			default:
				return err
			}
		}
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	securityGroup, warnings, err := cmd.Actor.GetSecurityGroupByName(cmd.RequiredArgs.SecurityGroupName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.OrganizationName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	spacesToBind := []v2action.Space{}
	if cmd.RequiredArgs.SpaceName != "" {
		var space v2action.Space
		space, warnings, err = cmd.Actor.GetSpaceByOrganizationAndName(org.GUID, cmd.RequiredArgs.SpaceName)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		spacesToBind = append(spacesToBind, space)
	} else {
		var spaces []v2action.Space
		spaces, warnings, err = cmd.Actor.GetOrganizationSpaces(org.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		spacesToBind = append(spacesToBind, spaces...)
	}

	for _, space := range spacesToBind {
		cmd.UI.DisplayTextWithFlavor("Assigning security group {{.security_group}} to space {{.space}} in org {{.organization}} as {{.username}}...", map[string]interface{}{
			"security_group": securityGroup.Name,
			"space":          space.Name,
			"organization":   org.Name,
			"username":       user.Name,
		})

		warnings, err = cmd.Actor.BindSecurityGroupToSpace(securityGroup.GUID, space.GUID, ccv2.SecurityGroupLifecycle(cmd.Lifecycle))
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
	}

	cmd.UI.DisplayText("TIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
