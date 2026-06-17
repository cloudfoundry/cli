package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/command/flag"
)

type AddRoutePolicyCommand struct {
	BaseCommand

	RequiredArgs flag.AddRoutePolicyArgs `positional-args:"yes"`
	Hostname     string                  `long:"hostname" required:"true" description:"Hostname for the route"`
	Path         string                  `long:"path" description:"Path for the route"`
	RoutePolicySourceFlags

	usage           interface{} `usage:"CF_NAME add-route-policy DOMAIN --hostname HOSTNAME [--source-app APP_NAME [--source-space SPACE_NAME] [--source-org ORG_NAME] | --source-space SPACE_NAME [--source-org ORG_NAME] | --source-org ORG_NAME | --source-any | --source SOURCE] [--path PATH]\n\nALLOW ACCESS TO A ROUTE:\n   Create a route policy that allows specific apps, spaces, or orgs to access a route using mTLS authentication.\n\nEXAMPLES:\n   # Allow the \"frontend-app\" (in current space) to access the backend route\n   cf add-route-policy apps.identity --source-app frontend-app --hostname backend\n\n   # Allow an app in a different space to access the route\n   cf add-route-policy apps.identity --source-app api-client --source-space other-space --hostname backend\n\n   # Allow an app in a different org to access the route\n   cf add-route-policy apps.identity --source-app external-client --source-space external-space --source-org external-org --hostname backend\n\n   # Allow all apps in the \"monitoring\" space to access the API metrics endpoint\n   cf add-route-policy apps.identity --source-space monitoring --hostname api --path /metrics\n\n   # Allow all apps in a space in a different org\n   cf add-route-policy apps.identity --source-space prod-space --source-org prod-org --hostname api\n\n   # Allow all apps in the \"platform\" org to access the route\n   cf add-route-policy apps.identity --source-org platform --hostname shared-api\n\n   # Allow any authenticated app to access the public API\n   cf add-route-policy apps.identity --source-any --hostname public-api\n\n   # Use raw source (advanced)\n   cf add-route-policy apps.identity --source cf:app:d76446a1-f429-4444-8797-be2f78b75b08 --hostname backend"`
	relatedCommands interface{} `related_commands:"route-policies, remove-route-policy, create-shared-domain"`
}

func (cmd AddRoutePolicyCommand) Execute(args []string) error {
	if err := cmd.RoutePolicySourceFlags.validateSourceFlags(); err != nil {
		return err
	}

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	source, scopeDisplay, warnings, err := resolveSource(cmd.RoutePolicySourceFlags, cmd.Actor, cmd.Config)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if err := validateSource(source); err != nil {
		return err
	}

	domainName := cmd.RequiredArgs.Domain

	cmd.UI.DisplayTextWithFlavor("Adding route policy for route {{.Hostname}}.{{.Domain}}{{.Path}} as {{.User}}...",
		map[string]interface{}{
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
			"User":     user.Name,
		})

	cmd.UI.DisplayText("  {{.ScopeDisplay}}",
		map[string]interface{}{
			"ScopeDisplay": scopeDisplay,
		})
	cmd.UI.DisplayText("  source: {{.Source}}",
		map[string]interface{}{
			"Source": source,
		})

	warnings, err = cmd.Actor.AddRoutePolicy(domainName, source, cmd.Hostname, cmd.Path)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}

func validateSource(source string) error {
	validPrefixes := []string{"cf:app:", "cf:space:", "cf:org:", "cf:any"}
	for _, prefix := range validPrefixes {
		if len(source) >= len(prefix) && source[:len(prefix)] == prefix {
			if prefix == "cf:any" {
				if source != "cf:any" {
					return fmt.Errorf("source 'cf:any' must not have a GUID suffix")
				}
				return nil
			}
			if len(source) <= len(prefix) {
				return fmt.Errorf("source '%s' must include a GUID (e.g., %s<guid>)", source, prefix)
			}
			return nil
		}
	}
	return fmt.Errorf("source must start with one of: cf:app:, cf:space:, cf:org:, or be exactly 'cf:any'")
}

func formatPath(path string) string {
	if path == "" {
		return ""
	}
	return path
}
