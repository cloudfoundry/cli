package v3

import (
	"net/http"

	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . AddNetworkPolicyActor

type AddNetworkPolicyActor interface {
	AddNetworkPolicy(spaceGUID string, srcAppName string, destAppName string, protocol string, startPort int, endPort int) (cfnetworkingaction.Warnings, error)
}

type AddNetworkPolicyCommand struct {
	RequiredArgs   flag.AddNetworkPolicyArgs `positional-args:"yes"`
	DestinationApp string                    `long:"destination-app" required:"true" description:"Name of app to connect to"`
	Port           flag.NetworkPort          `long:"port" description:"Port or range of ports for connection to destination app (Default: 8080)"`
	Protocol       flag.NetworkProtocol      `long:"protocol" description:"Protocol to connect apps with (Default: tcp)"`

	usage           interface{} `usage:"CF_NAME add-network-policy SOURCE_APP --destination-app DESTINATION_APP [(--protocol (tcp | udp) --port RANGE)]\n\nEXAMPLES:\n   CF_NAME add-network-policy frontend --destination-app backend --protocol tcp --port 8081\n   CF_NAME add-network-policy frontend --destination-app backend --protocol tcp --port 8080-8090"`
	relatedCommands interface{} `related_commands:"apps, network-policies"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       AddNetworkPolicyActor
}

func (cmd *AddNetworkPolicyCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, uaa, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.CFNetworkingEndpointNotFoundError{}
		}

		return err
	}

	v3Actor := v3action.NewActor(client, config, nil, nil)
	networkingClient, err := shared.NewNetworkingClient(client.NetworkPolicyV1(), config, uaa, ui)
	if err != nil {
		return err
	}
	cmd.Actor = cfnetworkingaction.NewActor(networkingClient, v3Actor)

	return nil
}

func (cmd AddNetworkPolicyCommand) Execute(args []string) error {
	switch {
	case cmd.Protocol.Protocol != "" && cmd.Port.StartPort == 0 && cmd.Port.EndPort == 0:
		return translatableerror.NetworkPolicyProtocolOrPortNotProvidedError{}
	case cmd.Protocol.Protocol == "" && (cmd.Port.StartPort != 0 || cmd.Port.EndPort != 0):
		return translatableerror.NetworkPolicyProtocolOrPortNotProvidedError{}
	case cmd.Protocol.Protocol == "" && cmd.Port.StartPort == 0 && cmd.Port.EndPort == 0:
		cmd.Protocol.Protocol = "tcp"
		cmd.Port.StartPort = 8080
		cmd.Port.EndPort = 8080
	}

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("Adding network policy to app {{.SrcAppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"SrcAppName": cmd.RequiredArgs.SourceApp,
		"Org":        cmd.Config.TargetedOrganization().Name,
		"Space":      cmd.Config.TargetedSpace().Name,
		"User":       user.Name,
	})

	warnings, err := cmd.Actor.AddNetworkPolicy(cmd.Config.TargetedSpace().GUID, cmd.RequiredArgs.SourceApp, cmd.DestinationApp, cmd.Protocol.Protocol, cmd.Port.StartPort, cmd.Port.EndPort)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
