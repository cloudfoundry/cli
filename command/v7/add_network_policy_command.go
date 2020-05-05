package v7

import (
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . NetworkingActor

type NetworkingActor interface {
	AddNetworkPolicy(srcSpaceGUID string, srcAppName string, destSpaceGUID string, destAppName string, protocol string, startPort int, endPort int) (cfnetworkingaction.Warnings, error)
}

type AddNetworkPolicyCommand struct {
	BaseCommand

	RequiredArgs flag.AddNetworkPolicyArgsV7 `positional-args:"yes"`
	Port         flag.NetworkPort            `long:"port" description:"Port or range of ports for connection to destination app (Default: 8080)"`
	Protocol     flag.NetworkProtocol        `long:"protocol" description:"Protocol to connect apps with (Default: tcp)"`

	DestinationOrg   string `short:"o" description:"The org of the destination app (Default: targeted org)"`
	DestinationSpace string `short:"s" description:"The space of the destination app (Default: targeted space)"`

	usage           interface{} `usage:"CF_NAME add-network-policy SOURCE_APP DESTINATION_APP [-s DESTINATION_SPACE_NAME [-o DESTINATION_ORG_NAME]] [--protocol (tcp | udp) --port RANGE]\n\nEXAMPLES:\n   CF_NAME add-network-policy frontend backend --protocol tcp --port 8081\n   CF_NAME add-network-policy frontend backend -s backend-space -o backend-org --protocol tcp --port 8080-8090"`
	relatedCommands interface{} `related_commands:"apps, network-policies, remove-network-policy"`

	NetworkingActor NetworkingActor
}

func (cmd *AddNetworkPolicyCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd AddNetworkPolicyCommand) Execute(args []string) error {
	switch {
	case cmd.Protocol.Protocol != "" && cmd.Port.StartPort == 0 && cmd.Port.EndPort == 0:
		return translatableerror.NetworkPolicyProtocolOrPortNotProvidedError{}
	case cmd.Protocol.Protocol == "" && (cmd.Port.StartPort != 0 || cmd.Port.EndPort != 0):
		return translatableerror.NetworkPolicyProtocolOrPortNotProvidedError{}
	case cmd.DestinationOrg != "" && cmd.DestinationSpace == "":
		return translatableerror.NetworkPolicyDestinationOrgWithoutSpaceError{}
	case cmd.Protocol.Protocol == "" && cmd.Port.StartPort == 0 && cmd.Port.EndPort == 0:
		cmd.Protocol.Protocol = "tcp"
		cmd.Port.StartPort = 8080
		cmd.Port.EndPort = 8080
	}

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	destOrgGUID := cmd.Config.TargetedOrganization().GUID

	displayDestinationOrg := cmd.Config.TargetedOrganization().Name
	if cmd.DestinationOrg != "" {
		var destOrg v7action.Organization
		var warnings v7action.Warnings
		destOrg, warnings, err = cmd.Actor.GetOrganizationByName(cmd.DestinationOrg)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		destOrgGUID = destOrg.GUID
		displayDestinationOrg = cmd.DestinationOrg
	}

	destSpaceGUID := cmd.Config.TargetedSpace().GUID

	if cmd.DestinationSpace != "" {
		var destSpace v7action.Space
		var warnings v7action.Warnings
		destSpace, warnings, err = cmd.Actor.GetSpaceByNameAndOrganization(cmd.DestinationSpace, destOrgGUID)
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
		cmd.UI.DisplayTextWithFlavor("Adding network policy from app {{.SrcAppName}} in org {{.Org}} / space {{.Space}} to app {{.DstAppName}} in org {{.DstOrg}} / space {{.DstSpace}} as {{.User}}...", map[string]interface{}{
			"SrcAppName": cmd.RequiredArgs.SourceApp,
			"Org":        cmd.Config.TargetedOrganization().Name,
			"Space":      cmd.Config.TargetedSpace().Name,
			"DstAppName": cmd.RequiredArgs.DestApp,
			"DstOrg":     displayDestinationOrg,
			"DstSpace":   cmd.DestinationSpace,
			"User":       user.Name,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Adding network policy from app {{.SrcAppName}} to app {{.DstAppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
			"SrcAppName": cmd.RequiredArgs.SourceApp,
			"Org":        cmd.Config.TargetedOrganization().Name,
			"Space":      cmd.Config.TargetedSpace().Name,
			"DstAppName": cmd.RequiredArgs.DestApp,
			"User":       user.Name,
		})
	}

	warnings, err := cmd.NetworkingActor.AddNetworkPolicy(cmd.Config.TargetedSpace().GUID, cmd.RequiredArgs.SourceApp, destSpaceGUID, cmd.RequiredArgs.DestApp, cmd.Protocol.Protocol, cmd.Port.StartPort, cmd.Port.EndPort)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
