package v3

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . ListNetworkAccessActor

type ListNetworkAccessActor interface {
	ListNetworkAccess(spaceGUID string, srcAppName string) ([]cfnetworkingaction.Policy, cfnetworkingaction.Warnings, error)
}

type ListNetworkAccessCommand struct {
	SourceApp string `long:"source" required:"false" description:"Source app to filter results by (optional)"`

	usage           interface{} `usage:"CF_NAME list-network-access [--source SOURCE_APP]"`
	relatedCommands interface{} `related_commands:"allow-network-access, apps, remove-network-access"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ListNetworkAccessActor
}

func (cmd *ListNetworkAccessCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, uaa, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	v3Actor := v3action.NewActor(client, config)
	networkingClient, err := shared.NewNetworkingClient(client.NetworkPolicyV1(), config, uaa, ui)
	if err != nil {
		return err
	}
	cmd.Actor = cfnetworkingaction.NewActor(networkingClient, v3Actor)

	return nil
}

func (cmd ListNetworkAccessCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayTextWithFlavor("Listing network traffic as {{.User}}...", map[string]interface{}{
		"User": user.Name,
	})

	policies, warnings, err := cmd.Actor.ListNetworkAccess(cmd.Config.TargetedSpace().GUID, cmd.SourceApp)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	table := [][]string{
		{
			cmd.UI.TranslateText("Source"),
			cmd.UI.TranslateText("Destination"),
			cmd.UI.TranslateText("Protocol"),
			cmd.UI.TranslateText("Ports"),
		},
	}

	for _, policy := range policies {
		table = append(table, []string{
			policy.SourceName,
			policy.DestinationName,
			policy.Protocol,
			fmt.Sprintf("%d-%d", policy.StartPort, policy.EndPort),
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, 3)

	return nil
}
