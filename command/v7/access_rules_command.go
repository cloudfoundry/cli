package v7

import (
	"code.cloudfoundry.org/cli/v9/util/ui"
)

type AccessRulesCommand struct {
	BaseCommand

	Domain   string `long:"domain" description:"Filter by domain name"`
	Hostname string `long:"hostname" description:"Filter by hostname"`
	Path     string `long:"path" description:"Filter by path"`
	Labels   string `long:"labels" description:"Selector to filter access rules by labels"`

	usage           interface{} `usage:"CF_NAME access-rules [--domain DOMAIN] [--hostname HOSTNAME] [--path PATH] [--labels SELECTOR]\n\nEXAMPLES:\n   cf access-rules\n   cf access-rules --domain apps.identity\n   cf access-rules --domain apps.identity --hostname backend\n   cf access-rules --labels env=prod"`
	relatedCommands interface{} `related_commands:"add-access-rule, remove-access-rule, routes"`
}

func (cmd AccessRulesCommand) Execute(args []string) error {
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
		"Getting access rules in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	// Fetch access rules for space with filters
	rulesWithRoutes, warnings, err := cmd.Actor.GetAccessRulesForSpace(
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
	if len(rulesWithRoutes) == 0 {
		cmd.UI.DisplayText("No access rules found.")
		return nil
	}

	// Build table data
	table := [][]string{
		{
			cmd.UI.TranslateText("host"),
			cmd.UI.TranslateText("domain"),
			cmd.UI.TranslateText("path"),
			cmd.UI.TranslateText("selector"),
			cmd.UI.TranslateText("scope"),
			cmd.UI.TranslateText("source"),
		},
	}

	for _, ruleWithRoute := range rulesWithRoutes {
		table = append(table, []string{
			ruleWithRoute.Route.Host,
			ruleWithRoute.DomainName,
			ruleWithRoute.Route.Path,
			ruleWithRoute.Selector,
			ruleWithRoute.ScopeType,
			ruleWithRoute.SourceName,
		})
	}

	// Display table
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)

	return nil
}

