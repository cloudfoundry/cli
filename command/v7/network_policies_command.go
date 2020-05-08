package v7

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . NetworkPoliciesActor

type NetworkPoliciesActor interface {
	NetworkPoliciesBySpaceAndAppName(spaceGUID string, srcAppName string) ([]cfnetworkingaction.Policy, cfnetworkingaction.Warnings, error)
	NetworkPoliciesBySpace(spaceGUID string) ([]cfnetworkingaction.Policy, cfnetworkingaction.Warnings, error)
}

type NetworkPoliciesCommand struct {
	command.BaseCommand

	SourceApp string `long:"source" required:"false" description:"Source app to filter results by"`

	usage           interface{} `usage:"CF_NAME network-policies [--source SOURCE_APP]"`
	relatedCommands interface{} `related_commands:"add-network-policy, apps, remove-network-policy"`

	NetworkingActor NetworkPoliciesActor
}

func (cmd *NetworkPoliciesCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd NetworkPoliciesCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	var policies []cfnetworkingaction.Policy
	var warnings cfnetworkingaction.Warnings

	if cmd.SourceApp != "" {
		cmd.UI.DisplayTextWithFlavor("Listing network policies of app {{.SrcAppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
			"SrcAppName": cmd.SourceApp,
			"Org":        cmd.Config.TargetedOrganization().Name,
			"Space":      cmd.Config.TargetedSpace().Name,
			"User":       user.Name,
		})
		policies, warnings, err = cmd.NetworkingActor.NetworkPoliciesBySpaceAndAppName(cmd.Config.TargetedSpace().GUID, cmd.SourceApp)
	} else {
		cmd.UI.DisplayTextWithFlavor("Listing network policies in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
			"Org":   cmd.Config.TargetedOrganization().Name,
			"Space": cmd.Config.TargetedSpace().Name,
			"User":  user.Name,
		})
		policies, warnings, err = cmd.NetworkingActor.NetworkPoliciesBySpace(cmd.Config.TargetedSpace().GUID)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()

	table := [][]string{
		{
			cmd.UI.TranslateText("source"),
			cmd.UI.TranslateText("destination"),
			cmd.UI.TranslateText("protocol"),
			cmd.UI.TranslateText("ports"),
			cmd.UI.TranslateText("destination space"),
			cmd.UI.TranslateText("destination org"),
		},
	}

	for _, policy := range policies {
		var portEntry string
		if policy.StartPort == policy.EndPort {
			portEntry = strconv.Itoa(policy.StartPort)
		} else {
			portEntry = fmt.Sprintf("%d-%d", policy.StartPort, policy.EndPort)
		}
		table = append(table, []string{
			policy.SourceName,
			policy.DestinationName,
			policy.Protocol,
			portEntry,
			policy.DestinationSpaceName,
			policy.DestinationOrgName,
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
