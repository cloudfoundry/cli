package v3

import (
	"fmt"
	"net/http"
	"strconv"

	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . NetworkPoliciesActor

type NetworkPoliciesActor interface {
	NetworkPoliciesBySpaceAndAppName(spaceGUID string, srcAppName string) ([]cfnetworkingaction.Policy, cfnetworkingaction.Warnings, error)
	NetworkPoliciesBySpace(spaceGUID string) ([]cfnetworkingaction.Policy, cfnetworkingaction.Warnings, error)
}

type NetworkPoliciesCommand struct {
	SourceApp string `long:"source" required:"false" description:"Source app to filter results by"`

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
		policies, warnings, err = cmd.Actor.NetworkPoliciesBySpaceAndAppName(cmd.Config.TargetedSpace().GUID, cmd.SourceApp)
	} else {
		cmd.UI.DisplayTextWithFlavor("Listing network policies in org {{.Org}} / space {{.Space}} as {{.User}}...", map[string]interface{}{
			"Org":   cmd.Config.TargetedOrganization().Name,
			"Space": cmd.Config.TargetedSpace().Name,
			"User":  user.Name,
		})
		policies, warnings, err = cmd.Actor.NetworkPoliciesBySpace(cmd.Config.TargetedSpace().GUID)
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
		})
	}

	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
