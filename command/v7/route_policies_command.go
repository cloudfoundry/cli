package v7

import (
	"code.cloudfoundry.org/cli/v9/util/ui"
)

type RoutePoliciesCommand struct {
	BaseCommand

	Domain   string `long:"domain" description:"Filter by domain name"`
	Hostname string `long:"hostname" description:"Filter by hostname"`
	Path     string `long:"path" description:"Filter by path"`
	Labels   string `long:"labels" description:"Selector to filter route policies by labels"`

	usage           interface{} `usage:"CF_NAME route-policies [--domain DOMAIN] [--hostname HOSTNAME] [--path PATH] [--labels SELECTOR]\n\nEXAMPLES:\n   cf route-policies\n   cf route-policies --domain apps.identity\n   cf route-policies --domain apps.identity --hostname backend\n   cf route-policies --labels env=prod"`
	relatedCommands interface{} `related_commands:"add-route-policy, remove-route-policy, routes"`
}

func (cmd RoutePoliciesCommand) Execute(args []string) error {
	// Check target (org + space required)
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	// Get current user
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	// Display contextual header
	cmd.UI.DisplayTextWithFlavor(
		"Getting route policies in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	// Fetch route policies for space with filters
	policiesWithRoutes, warnings, err := cmd.Actor.GetRoutePoliciesForSpace(
		cmd.Config.TargetedSpace().GUID,
		cmd.Domain,
		cmd.Hostname,
		cmd.Path,
		cmd.Labels,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	// Handle empty results
	if len(policiesWithRoutes) == 0 {
		cmd.UI.DisplayText("No route policies found.")
		return nil
	}

	// Build table data
	table := [][]string{
		{
			cmd.UI.TranslateText("host"),
			cmd.UI.TranslateText("domain"),
			cmd.UI.TranslateText("path"),
			cmd.UI.TranslateText("source"),
			cmd.UI.TranslateText("scope"),
			cmd.UI.TranslateText("name"),
		},
	}

	for _, policyWithRoute := range policiesWithRoutes {
		table = append(table, []string{
			policyWithRoute.Route.Host,
			policyWithRoute.DomainName,
			policyWithRoute.Route.Path,
			policyWithRoute.Source,
			policyWithRoute.ScopeType,
			policyWithRoute.SourceName,
		})
	}

	// Display table
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}
