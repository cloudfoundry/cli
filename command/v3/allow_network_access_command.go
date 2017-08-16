package v3

import (
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . AllowNetworkAccessActor

type AllowNetworkAccessActor interface {
	AllowNetworkAccess(spaceGUID string, srcAppName string, destAppName string, protocol string, startPort int, endPort int) (cfnetworkingaction.Warnings, error)
}

type AllowNetworkAccessCommand struct {
	RequiredArgs   flag.AllowNetworkAccessArgs `positional-args:"yes"`
	DestinationApp string                      `long:"destination-app" required:"true" description:"The destination app"`
	Port           flag.NetworkPort            `long:"port" description:"Port or range to connect to destination app with" default:"8080"`
	Protocol       flag.NetworkProtocol        `long:"protocol" description:"Protocol to connect apps with" default:"tcp"`

	usage           interface{} `usage:"CF_NAME allow-network-access SOURCE_APP --destination-app DESTINATION_APP [(--protocol (tcp | udp) --port RANGE)]\n\nEXAMPLES:\n   CF_NAME allow-network-access frontend --destination-app backend --protocol tcp --port 8081\n   CF_NAME allow-network-access frontend --destination-app backend --protocol tcp --port 8080-8090"`
	relatedCommands interface{} `related_commands:"apps, list-network-access"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       AllowNetworkAccessActor
}

func (cmd *AllowNetworkAccessCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, uaa, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v3Actor := v3action.NewActor(client, config)
	networkingClient := shared.NewNetworkingClient(client.NetworkPolicyV1(), config, uaa, ui)
	cmd.Actor = cfnetworkingaction.NewActor(networkingClient, v3Actor)

	return nil
}

func (cmd AllowNetworkAccessCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("Allowing network traffic from app {{.SrcAppName}} to {{.DestAppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
		"SrcAppName":  cmd.RequiredArgs.SourceApp,
		"DestAppName": cmd.DestinationApp,
		"Org":         cmd.Config.TargetedOrganization().Name,
		"Space":       cmd.Config.TargetedSpace().Name,
		"User":        user.Name,
	})

	warnings, err := cmd.Actor.AllowNetworkAccess(cmd.Config.TargetedSpace().GUID, cmd.RequiredArgs.SourceApp, cmd.DestinationApp, cmd.Protocol.Protocol, cmd.Port.StartPort, cmd.Port.EndPort)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}
	cmd.UI.DisplayOK()

	return nil
}
