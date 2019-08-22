package v6

import (
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . AddNetworkPolicyActor

type AddNetworkPolicyActor interface {
	AddNetworkPolicy(srcSpaceGUID string, srcAppName string, destSpaceGUID string, destAppName string, protocol string, startPort int, endPort int) (cfnetworkingaction.Warnings, error)
}

//go:generate counterfeiter . MembershipActor

type MembershipActor interface {
	GetOrganizationByName(name string) (v3action.Organization, v3action.Warnings, error)
	GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (v3action.Space, v3action.Warnings, error)
}

type AddNetworkPolicyCommand struct {
	RequiredArgs   flag.AddNetworkPolicyArgs `positional-args:"yes"`
	DestinationApp string                    `long:"destination-app" required:"true" description:"Name of app to connect to"`
	Port           flag.NetworkPort          `long:"port" description:"Port or range of ports for connection to destination app (Default: 8080)"`
	Protocol       flag.NetworkProtocol      `long:"protocol" description:"Protocol to connect apps with (Default: tcp)"`

	DestinationOrg   string `short:"o" description:"The org of the destination app (Default: targeted org)"`
	DestinationSpace string `short:"s" description:"The space of the destination app (Default: targeted space)"`

	usage           interface{} `usage:"CF_NAME add-network-policy SOURCE_APP --destination-app DESTINATION_APP [-s DESTINATION_SPACE_NAME [-o DESTINATION_ORG_NAME]] [--protocol (tcp | udp) --port RANGE]\n\nEXAMPLES:\n   CF_NAME add-network-policy frontend --destination-app backend --protocol tcp --port 8081\n   CF_NAME add-network-policy frontend --destination-app backend -s backend-space -o backend-org --protocol tcp --port 8080-8090"`
	relatedCommands interface{} `related_commands:"apps, network-policies, remove-network-policy"`

	UI                 command.UI
	Config             command.Config
	SharedActor        command.SharedActor
	NetworkPolicyActor AddNetworkPolicyActor
	MembershipActor    MembershipActor
}

func (cmd *AddNetworkPolicyCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, uaa, err := shared.NewV3BasedClients(config, ui, true)
	if err != nil {
		return err
	}

	v3Actor := v3action.NewActor(client, config, nil, nil)

	networkingClient, err := shared.NewNetworkingClient(client.NetworkPolicyV1(), config, uaa, ui)
	if err != nil {
		return err
	}
	cmd.NetworkPolicyActor = cfnetworkingaction.NewActor(networkingClient, v3Actor)
	cmd.MembershipActor = v3Actor

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
		var destOrg v3action.Organization
		var warnings v3action.Warnings
		destOrg, warnings, err = cmd.MembershipActor.GetOrganizationByName(cmd.DestinationOrg)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		destOrgGUID = destOrg.GUID
		displayDestinationOrg = cmd.DestinationOrg
	}

	destSpaceGUID := cmd.Config.TargetedSpace().GUID

	if cmd.DestinationSpace != "" {
		var destSpace v3action.Space
		var warnings v3action.Warnings
		destSpace, warnings, err = cmd.MembershipActor.GetSpaceByNameAndOrganization(cmd.DestinationSpace, destOrgGUID)
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
			"DstAppName": cmd.DestinationApp,
			"DstOrg":     displayDestinationOrg,
			"DstSpace":   cmd.DestinationSpace,
			"User":       user.Name,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Adding network policy from app {{.SrcAppName}} to app {{.DstAppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
			"SrcAppName": cmd.RequiredArgs.SourceApp,
			"Org":        cmd.Config.TargetedOrganization().Name,
			"Space":      cmd.Config.TargetedSpace().Name,
			"DstAppName": cmd.DestinationApp,
			"User":       user.Name,
		})
	}

	warnings, err := cmd.NetworkPolicyActor.AddNetworkPolicy(cmd.Config.TargetedSpace().GUID, cmd.RequiredArgs.SourceApp, destSpaceGUID, cmd.DestinationApp, cmd.Protocol.Protocol, cmd.Port.StartPort, cmd.Port.EndPort)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
