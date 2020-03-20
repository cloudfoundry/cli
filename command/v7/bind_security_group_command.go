package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . BindSecurityGroupActor

type BindSecurityGroupActor interface {
	BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string, lifecycle constant.SecurityGroupLifecycle) (v7action.Warnings, error)
	CloudControllerAPIVersion() string
	GetOrganizationByName(orgName string) (v7action.Organization, v7action.Warnings, error)
	GetOrganizationSpaces(orgGUID string) ([]v7action.Space, v7action.Warnings, error)
	GetSecurityGroup(securityGroupName string) (resources.SecurityGroup, v7action.Warnings, error)
	GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (v7action.Space, v7action.Warnings, error)
}

type BindSecurityGroupCommand struct {
	RequiredArgs    flag.BindSecurityGroupV7Args `positional-args:"yes"`
	Lifecycle       flag.SecurityGroupLifecycle  `long:"lifecycle" choice:"running" choice:"staging" default:"running" description:"Lifecycle phase the group applies to."`
	Space           string                       `long:"space" description:"Space to bind the security group to. (Default: all existing spaces in org)"`
	usage           interface{}                  `usage:"CF_NAME bind-security-group SECURITY_GROUP ORG [--lifecycle (running | staging)] [--space SPACE]\n\nTIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications."`
	relatedCommands interface{}                  `related_commands:"apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       BindSecurityGroupActor
}

func (cmd *BindSecurityGroupCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd BindSecurityGroupCommand) Execute(args []string) error {
	var err error

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	securityGroup, warnings, err := cmd.Actor.GetSecurityGroup(cmd.RequiredArgs.SecurityGroupName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.RequiredArgs.OrganizationName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	spacesToBind := []v7action.Space{}
	if cmd.Space != "" {
		var space v7action.Space
		space, warnings, err = cmd.Actor.GetSpaceByNameAndOrganization(cmd.Space, org.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		spacesToBind = append(spacesToBind, space)
	} else {
		var spaces []v7action.Space
		spaces, warnings, err = cmd.Actor.GetOrganizationSpaces(org.GUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		spacesToBind = append(spacesToBind, spaces...)
	}

	for _, space := range spacesToBind {
		cmd.UI.DisplayTextWithFlavor("Assigning {{.lifecycle}} security group {{.security_group}} to space {{.space}} in org {{.organization}} as {{.username}}...", map[string]interface{}{
			"lifecycle":      constant.SecurityGroupLifecycle(cmd.Lifecycle),
			"security_group": securityGroup.Name,
			"space":          space.Name,
			"organization":   org.Name,
			"username":       user.Name,
		})

		warnings, err = cmd.Actor.BindSecurityGroupToSpace(securityGroup.GUID, space.GUID, constant.SecurityGroupLifecycle(cmd.Lifecycle))
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
	}

	cmd.UI.DisplayText("TIP: Changes require an app restart (for running) or restage (for staging) to apply to existing applications.")

	return nil
}
