package v3

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . NetworkPoliciesActor

type NetworkPoliciesActor interface {
	NetworkPoliciesBySpaceAndAppName(spaceGUID string, srcAppName string) ([]cfnetworkingaction.Policy, cfnetworkingaction.Warnings, error)
	NetworkPoliciesBySpace(spaceGUID string) ([]cfnetworkingaction.Policy, cfnetworkingaction.Warnings, error)
}

type NetworkPoliciesCommand struct {
	SourceApp string `long:"source" required:"false" description:"Source app to filter results by (optional)"`

	usage           interface{} `usage:"CF_NAME network-policies [--source SOURCE_APP]"`
	relatedCommands interface{} `related_commands:"add-network-policy, apps, remove-network-policy"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       NetworkPoliciesActor
}

func (cmd *NetworkPoliciesCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd NetworkPoliciesCommand) Execute(args []string) error {
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

	var policies []cfnetworkingaction.Policy
	var warnings cfnetworkingaction.Warnings

	if cmd.SourceApp != "" {
		policies, warnings, err = cmd.Actor.NetworkPoliciesBySpaceAndAppName(cmd.Config.TargetedSpace().GUID, cmd.SourceApp)
	} else {
		policies, warnings, err = cmd.Actor.NetworkPoliciesBySpace(cmd.Config.TargetedSpace().GUID)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayNewline()

	table := [][]string{
		{
			cmd.UI.TranslateText("source"),
			cmd.UI.TranslateText("destination"),
			cmd.UI.TranslateText("protocol"),
			cmd.UI.TranslateText("ports"),
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
