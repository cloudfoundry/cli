package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . RemoveNetworkPolicyActor

type RemoveNetworkPolicyActor interface {
	RemoveNetworkPolicy(srcSpaceGUID string, srcAppName string, destSpaceGUID string, destAppName string, protocol string, startPort int, endPort int) (cfnetworkingaction.Warnings, error)
}

type RemoveNetworkPolicyCommand struct {
	RequiredArgs     flag.RemoveNetworkPolicyArgs `positional-args:"yes"`
	DestinationApp   string                       `long:"destination-app" required:"true" description:"Name of app to connect to"`
	Port             flag.NetworkPort             `long:"port" required:"true" description:"Port or range of ports that destination app is connected with"`
	Protocol         flag.NetworkProtocol         `long:"protocol" required:"true" description:"Protocol that apps are connected with"`
	DestinationOrg   string                       `short:"o" description:"The org of the destination app (Default: targeted org)"`
	DestinationSpace string                       `short:"s" description:"The space of the destination app (Default: targeted space)"`

	usage           interface{} `usage:"CF_NAME remove-network-policy SOURCE_APP --destination-app DESTINATION_APP [-s DESTINATION_SPACE_NAME [-o DESTINATION_ORG_NAME]] --protocol (tcp | udp) --port RANGE\n\nEXAMPLES:\n   CF_NAME remove-network-policy frontend --destination-app backend --protocol tcp --port 8081\n   CF_NAME remove-network-policy frontend --destination-app backend -s backend-space -o backend-org --protocol tcp --port 8080-8090"`
	relatedCommands interface{} `related_commands:"apps, network-policies, add-network-policy"`

	UI              command.UI
	Config          command.Config
	SharedActor     command.SharedActor
	NetworkingActor RemoveNetworkPolicyActor
	Actor           Actor
}

func (cmd *RemoveNetworkPolicyCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	actor := v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	cmd.Actor = actor

	networkingClient, err := shared.NewNetworkingClient(ccClient.NetworkPolicyV1(), config, uaaClient, ui)
	if err != nil {
		return err
	}
	cmd.NetworkingActor = cfnetworkingaction.NewActor(networkingClient, ccClient)

	return nil
}

func (cmd RemoveNetworkPolicyCommand) Execute(args []string) error {
	if cmd.DestinationOrg != "" && cmd.DestinationSpace == "" {
		return translatableerror.NetworkPolicyDestinationOrgWithoutSpaceError{}
	}

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	destOrgGUID := cmd.Config.TargetedOrganization().GUID
	displayDestinationOrg := cmd.Config.TargetedOrganization().Name
	if cmd.DestinationOrg != "" {
		destOrg, warnings, err := cmd.Actor.GetOrganizationByName(cmd.DestinationOrg)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		destOrgGUID = destOrg.GUID
		displayDestinationOrg = cmd.DestinationOrg
	}

	destSpaceGUID := cmd.Config.TargetedSpace().GUID
	if cmd.DestinationSpace != "" {
		destSpace, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.DestinationSpace, destOrgGUID)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		destSpaceGUID = destSpace.GUID
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if cmd.DestinationSpace != "" {
		cmd.UI.DisplayTextWithFlavor("Removing network policy from app {{.SrcAppName}} in org {{.Org}} / space {{.Space}} to app {{.DstAppName}} in org {{.DstOrg}} / space {{.DstSpace}} as {{.User}}...", map[string]interface{}{
			"SrcAppName": cmd.RequiredArgs.SourceApp,
			"Org":        cmd.Config.TargetedOrganization().Name,
			"Space":      cmd.Config.TargetedSpace().Name,
			"DstAppName": cmd.DestinationApp,
			"DstOrg":     displayDestinationOrg,
			"DstSpace":   cmd.DestinationSpace,
			"User":       user.Name,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Removing network policy from app {{.SrcAppName}} to app {{.DstAppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
			"SrcAppName": cmd.RequiredArgs.SourceApp,
			"Org":        cmd.Config.TargetedOrganization().Name,
			"Space":      cmd.Config.TargetedSpace().Name,
			"DstAppName": cmd.DestinationApp,
			"User":       user.Name,
		})
	}

	removeWarnings, err := cmd.NetworkingActor.RemoveNetworkPolicy(cmd.Config.TargetedSpace().GUID, cmd.RequiredArgs.SourceApp, destSpaceGUID, cmd.DestinationApp, cmd.Protocol.Protocol, cmd.Port.StartPort, cmd.Port.EndPort)
	cmd.UI.DisplayWarnings(removeWarnings)
	if err != nil {
		switch err.(type) {
		case actionerror.PolicyDoesNotExistError:
			cmd.UI.DisplayText("Policy does not exist.")
		default:
			return err
		}
	}
	cmd.UI.DisplayOK()

	return nil
}
