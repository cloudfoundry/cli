package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
)

type AddAccessRuleCommand struct {
	BaseCommand

	RequiredArgs flag.AddAccessRuleArgs `positional-args:"yes"`
	Hostname     string                 `long:"hostname" required:"true" description:"Hostname for the route"`
	Path         string                 `long:"path" description:"Path for the route"`

	// Source resolution flags (mutually exclusive as primary source)
	SourceApp   string `long:"source-app" description:"Allow access from this app (by name)"`
	SourceSpace string `long:"source-space" description:"Allow access from all apps in this space (by name) or specify the space for --source-app"`
	SourceOrg   string `long:"source-org" description:"Allow access from all apps in this org (by name) or specify the org for --source-space/--source-app"`
	SourceAny   bool   `long:"source-any" description:"Allow access from any authenticated app"`

	// Advanced: raw selector flag
	Selector string `long:"selector" description:"Raw selector (cf:app:<guid>, cf:space:<guid>, cf:org:<guid>, or cf:any)"`

	usage        interface{} `usage:"CF_NAME add-access-rule RULE_NAME DOMAIN --hostname HOSTNAME [--source-app APP_NAME [--source-space SPACE_NAME] [--source-org ORG_NAME] | --source-space SPACE_NAME [--source-org ORG_NAME] | --source-org ORG_NAME | --source-any | --selector SELECTOR] [--path PATH]\n\nALLOW ACCESS TO A ROUTE:\n   Create an access rule that allows specific apps, spaces, or orgs to access a route using mTLS authentication.\n\nEXAMPLES:\n   # Allow the \"frontend-app\" (in current space) to access the backend route\n   cf add-access-rule allow-frontend apps.identity --source-app frontend-app --hostname backend\n\n   # Allow an app in a different space to access the route\n   cf add-access-rule allow-other-space apps.identity --source-app api-client --source-space other-space --hostname backend\n\n   # Allow an app in a different org to access the route\n   cf add-access-rule allow-other-org apps.identity --source-app external-client --source-space external-space --source-org external-org --hostname backend\n\n   # Allow all apps in the \"monitoring\" space to access the API metrics endpoint\n   cf add-access-rule allow-monitoring apps.identity --source-space monitoring --hostname api --path /metrics\n\n   # Allow all apps in a space in a different org\n   cf add-access-rule allow-prod-space apps.identity --source-space prod-space --source-org prod-org --hostname api\n\n   # Allow all apps in the \"platform\" org to access the route\n   cf add-access-rule allow-platform-org apps.identity --source-org platform --hostname shared-api\n\n   # Allow any authenticated app to access the public API\n   cf add-access-rule allow-all apps.identity --source-any --hostname public-api\n\n   # Use raw selector (advanced)\n   cf add-access-rule allow-raw apps.identity --selector cf:app:d76446a1-f429-4444-8797-be2f78b75b08 --hostname backend"`
	relatedCommands interface{} `related_commands:"access-rules, remove-access-rule, create-shared-domain"`
}

func (cmd AddAccessRuleCommand) Execute(args []string) error {
	// Validate source flags
	if err := cmd.validateSourceFlags(); err != nil {
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

	// Resolve selector from source flags
	selector, scopeDisplay, warnings, err := cmd.resolveSelector()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	// Validate selector format
	if err := validateSelector(selector); err != nil {
		return err
	}

	ruleName := cmd.RequiredArgs.RuleName
	domainName := cmd.RequiredArgs.Domain

	cmd.UI.DisplayTextWithFlavor("Adding access rule {{.RuleName}} for route {{.Hostname}}.{{.Domain}}{{.Path}} as {{.User}}...",
		map[string]interface{}{
			"RuleName": ruleName,
			"Hostname": cmd.Hostname,
			"Domain":   domainName,
			"Path":     formatPath(cmd.Path),
			"User":     user.Name,
		})

	// Display resolved source (for transparency)
	cmd.UI.DisplayText("  {{.ScopeDisplay}}",
		map[string]interface{}{
			"ScopeDisplay": scopeDisplay,
		})
	cmd.UI.DisplayText("  selector: {{.Selector}}",
		map[string]interface{}{
			"Selector": selector,
		})

	warnings, err = cmd.Actor.AddAccessRule(ruleName, domainName, selector, cmd.Hostname, cmd.Path)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Access rule '{{.RuleName}}' has been created. It may take a few seconds for the rule to propagate to GoRouter.",
		map[string]interface{}{
			"RuleName": ruleName,
		})

	return nil
}

// validateSourceFlags ensures exactly one source target is specified and validates combinations
func (cmd AddAccessRuleCommand) validateSourceFlags() error {
	sourceFlags := []string{}

	if cmd.Selector != "" {
		sourceFlags = append(sourceFlags, "--selector")
	}
	if cmd.SourceApp != "" {
		sourceFlags = append(sourceFlags, "--source-app")
	}
	if cmd.SourceSpace != "" && cmd.SourceApp == "" {
		// --source-space only counts as a primary source if --source-app is NOT provided
		sourceFlags = append(sourceFlags, "--source-space")
	}
	if cmd.SourceOrg != "" && cmd.SourceSpace == "" && cmd.SourceApp == "" {
		// --source-org only counts as a primary source if neither --source-space nor --source-app are provided
		sourceFlags = append(sourceFlags, "--source-org")
	}
	if cmd.SourceAny {
		sourceFlags = append(sourceFlags, "--source-any")
	}

	if len(sourceFlags) == 0 {
		return translatableerror.RequiredArgumentError{
			ArgumentName: "one of: --source-app, --source-space, --source-org, --source-any, or --selector",
		}
	}

	if len(sourceFlags) > 1 {
		return translatableerror.ArgumentCombinationError{
			Args: sourceFlags,
		}
	}

	return nil
}

// resolveSelector resolves source flags to a selector string
// Returns (selector, scopeDisplay, warnings, error)
// scopeDisplay is a human-readable description for output (e.g., "scope: app, source: frontend-app")
func (cmd AddAccessRuleCommand) resolveSelector() (string, string, v7action.Warnings, error) {
	var allWarnings v7action.Warnings

	// Priority: --selector flag (raw selector, no resolution needed)
	if cmd.Selector != "" {
		return cmd.Selector, fmt.Sprintf("selector: %s", cmd.Selector), allWarnings, nil
	}

	// --source-any
	if cmd.SourceAny {
		return "cf:any", "scope: any, source: any authenticated app", allWarnings, nil
	}

	// --source-app (with optional --source-space and --source-org for cross-space/org lookup)
	if cmd.SourceApp != "" {
		// Determine space GUID for app lookup
		spaceGUID := cmd.Config.TargetedSpace().GUID
		spaceName := cmd.Config.TargetedSpace().Name
		orgName := cmd.Config.TargetedOrganization().Name

		if cmd.SourceSpace != "" {
			// Determine org GUID for space lookup
			orgGUID := cmd.Config.TargetedOrganization().GUID
			if cmd.SourceOrg != "" {
				org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.SourceOrg)
				allWarnings = append(allWarnings, warnings...)
				if err != nil {
					return "", "", allWarnings, err
				}
				orgGUID = org.GUID
				orgName = cmd.SourceOrg
			}

			// Resolve space by name
			space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.SourceSpace, orgGUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return "", "", allWarnings, err
			}
			spaceGUID = space.GUID
			spaceName = cmd.SourceSpace
		}

		// Resolve app by name in the determined space
		app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.SourceApp, spaceGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			// Enhanced error message for app not found
			if _, ok := err.(actionerror.ApplicationNotFoundError); ok {
				if cmd.SourceSpace == "" {
					// App not found in current space
					return "", "", allWarnings, fmt.Errorf(
						"App '%s' not found in space '%s' / org '%s'.\nTIP: If the app is in a different space or org, use --source-space and/or --source-org flags.",
						cmd.SourceApp,
						cmd.Config.TargetedSpace().Name,
						cmd.Config.TargetedOrganization().Name,
					)
				}
			}
			return "", "", allWarnings, err
		}

		scopeDisplay := fmt.Sprintf("scope: app, source: %s", cmd.SourceApp)
		if cmd.SourceSpace != "" {
			scopeDisplay += fmt.Sprintf(" (space: %s", spaceName)
			if cmd.SourceOrg != "" {
				scopeDisplay += fmt.Sprintf(", org: %s", orgName)
			}
			scopeDisplay += ")"
		}

		return fmt.Sprintf("cf:app:%s", app.GUID), scopeDisplay, allWarnings, nil
	}

	// --source-space (without --source-app, so create space-level rule)
	if cmd.SourceSpace != "" {
		// Determine org GUID for space lookup
		orgGUID := cmd.Config.TargetedOrganization().GUID
		orgName := cmd.Config.TargetedOrganization().Name
		if cmd.SourceOrg != "" {
			org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.SourceOrg)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return "", "", allWarnings, err
			}
			orgGUID = org.GUID
			orgName = cmd.SourceOrg
		}

		// Resolve space by name
		space, warnings, err := cmd.Actor.GetSpaceByNameAndOrganization(cmd.SourceSpace, orgGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", allWarnings, err
		}

		scopeDisplay := fmt.Sprintf("scope: space, source: %s", cmd.SourceSpace)
		if cmd.SourceOrg != "" {
			scopeDisplay += fmt.Sprintf(" (org: %s)", orgName)
		}

		return fmt.Sprintf("cf:space:%s", space.GUID), scopeDisplay, allWarnings, nil
	}

	// --source-org (without --source-space or --source-app, so create org-level rule)
	if cmd.SourceOrg != "" {
		org, warnings, err := cmd.Actor.GetOrganizationByName(cmd.SourceOrg)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", allWarnings, err
		}

		scopeDisplay := fmt.Sprintf("scope: org, source: %s", cmd.SourceOrg)

		return fmt.Sprintf("cf:org:%s", org.GUID), scopeDisplay, allWarnings, nil
	}

	// Should never reach here due to validation
	return "", "", allWarnings, fmt.Errorf("no source specified")
}

func validateSelector(selector string) error {
	// Basic validation - check for cf:app:, cf:space:, cf:org:, or cf:any prefix
	validPrefixes := []string{"cf:app:", "cf:space:", "cf:org:", "cf:any"}
	for _, prefix := range validPrefixes {
		if len(selector) >= len(prefix) && selector[:len(prefix)] == prefix {
			if prefix == "cf:any" {
				if selector != "cf:any" {
					return fmt.Errorf("selector 'cf:any' must not have a GUID suffix")
				}
				return nil
			}
			// For other selectors, ensure there's a GUID after the prefix
			if len(selector) <= len(prefix) {
				return fmt.Errorf("selector '%s' must include a GUID (e.g., %s<guid>)", selector, prefix)
			}
			return nil
		}
	}
	return fmt.Errorf("selector must start with one of: cf:app:, cf:space:, cf:org:, or be exactly 'cf:any'")
}

func formatPath(path string) string {
	if path == "" {
		return ""
	}
	return path
}
