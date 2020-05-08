package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . RemoveNetworkPolicyActor

type RemoveNetworkPolicyActor interface {
	RemoveNetworkPolicy(srcSpaceGUID string, srcAppName string, destSpaceGUID string, destAppName string, protocol string, startPort int, endPort int) (cfnetworkingaction.Warnings, error)
}

type RemoveNetworkPolicyCommand struct {
	command.BaseCommand

	RequiredArgs     flag.RemoveNetworkPolicyArgsV7 `positional-args:"yes"`
	Port             flag.NetworkPort               `long:"port" required:"true" description:"Port or range of ports that destination app is connected with"`
	Protocol         flag.NetworkProtocol           `long:"protocol" required:"true" description:"Protocol that apps are connected with"`
	DestinationOrg   string                         `short:"o" description:"The org of the destination app (Default: targeted org)"`
	DestinationSpace string                         `short:"s" description:"The space of the destination app (Default: targeted space)"`

	usage           interface{} `usage:"CF_NAME remove-network-policy SOURCE_APP DESTINATION_APP [-s DESTINATION_SPACE_NAME [-o DESTINATION_ORG_NAME]] --protocol (tcp | udp) --port RANGE\n\nEXAMPLES:\n   CF_NAME remove-network-policy frontend backend --protocol tcp --port 8081\n   CF_NAME remove-network-policy frontend backend -s backend-space -o backend-org --protocol tcp --port 8080-8090"`
	relatedCommands interface{} `related_commands:"apps, network-policies, add-network-policy"`

	NetworkingActor RemoveNetworkPolicyActor
}

func (cmd *RemoveNetworkPolicyCommand) Setup(config command.Config, ui command.UI) error {
	err := cmd.BaseCommand.Setup(config, ui)
	if err != nil {
		return err
	}

	ccClient, uaaClient := cmd.BaseCommand.GetClients()

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
			"DstAppName": cmd.RequiredArgs.DestApp,
			"DstOrg":     displayDestinationOrg,
			"DstSpace":   cmd.DestinationSpace,
			"User":       user.Name,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Removing network policy from app {{.SrcAppName}} to app {{.DstAppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
			"SrcAppName": cmd.RequiredArgs.SourceApp,
			"Org":        cmd.Config.TargetedOrganization().Name,
			"Space":      cmd.Config.TargetedSpace().Name,
			"DstAppName": cmd.RequiredArgs.DestApp,
			"User":       user.Name,
		})
	}

	removeWarnings, err := cmd.NetworkingActor.RemoveNetworkPolicy(cmd.Config.TargetedSpace().GUID, cmd.RequiredArgs.SourceApp, destSpaceGUID, cmd.RequiredArgs.DestApp, cmd.Protocol.Protocol, cmd.Port.StartPort, cmd.Port.EndPort)
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
