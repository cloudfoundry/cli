package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
)

// RoutePolicySourceFlags holds the mutually exclusive source-resolution flags
// shared between add-route-policy and remove-route-policy.
type RoutePolicySourceFlags struct {
	SourceApp   string `long:"source-app" description:"App name to identify the source (resolves to GUID)"`
	SourceSpace string `long:"source-space" description:"Space name to identify the source (by name) or specify the space for --source-app"`
	SourceOrg   string `long:"source-org" description:"Org name to identify the source (by name) or specify the org for --source-space/--source-app"`
	SourceAny   bool   `long:"source-any" description:"Any authenticated app"`
	Source      string `long:"source" description:"Raw source (cf:app:<guid>, cf:space:<guid>, cf:org:<guid>, or cf:any)"`
}

// validateSourceFlags ensures exactly one primary source is specified.
func (f RoutePolicySourceFlags) validateSourceFlags() error {
	sourceFlags := []string{}

	if f.Source != "" {
		sourceFlags = append(sourceFlags, "--source")
	}
	if f.SourceApp != "" {
		sourceFlags = append(sourceFlags, "--source-app")
	}
	if f.SourceSpace != "" && f.SourceApp == "" {
		sourceFlags = append(sourceFlags, "--source-space")
	}
	if f.SourceOrg != "" && f.SourceSpace == "" && f.SourceApp == "" {
		sourceFlags = append(sourceFlags, "--source-org")
	}
	if f.SourceAny {
		sourceFlags = append(sourceFlags, "--source-any")
	}

	// --source-org requires --source-space when used with --source-app
	if f.SourceOrg != "" && f.SourceApp != "" && f.SourceSpace == "" {
		return translatableerror.RequiredFlagsError{
			Arg1: "--source-org",
			Arg2: "--source-space",
		}
	}

	if len(sourceFlags) == 0 {
		return translatableerror.RequiredArgumentError{
			ArgumentName: "one of: --source-app, --source-space, --source-org, --source-any, or --source",
		}
	}

	if len(sourceFlags) > 1 {
		return translatableerror.ArgumentCombinationError{
			Args: sourceFlags,
		}
	}

	return nil
}

// resolveSource resolves source flags to a raw source string and a human-readable
// scope description. Returns (source, scopeDisplay, warnings, error).
//
// Resolution cascades org -> space -> app: each level refines the GUID that the
// next level resolves against, and the most specific flag provided becomes the
// scope. The result is always of the form cf:<scope>:<guid>.
func resolveSource(f RoutePolicySourceFlags, actor Actor, config command.Config) (string, string, v7action.Warnings, error) {
	var allWarnings v7action.Warnings

	// --source: raw value, no resolution needed
	if f.Source != "" {
		return f.Source, fmt.Sprintf("source: %s", f.Source), allWarnings, nil
	}

	// --source-any
	if f.SourceAny {
		return "cf:any", "scope: any, source: any authenticated app", allWarnings, nil
	}

	scope := ""
	var sourceGUID string

	// Org level: default to the targeted org; --source-org overrides it and, when it
	// is the most specific flag provided, becomes the scope.
	orgGUID := config.TargetedOrganization().GUID
	orgName := config.TargetedOrganization().Name
	if f.SourceOrg != "" {
		org, warnings, err := actor.GetOrganizationByName(f.SourceOrg)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", allWarnings, err
		}
		orgGUID, orgName = org.GUID, f.SourceOrg
		scope, sourceGUID = "org", orgGUID
	}

	// Space level: resolved within orgGUID (targeted or --source-org).
	spaceGUID := config.TargetedSpace().GUID
	spaceName := config.TargetedSpace().Name
	if f.SourceSpace != "" {
		space, warnings, err := actor.GetSpaceByNameAndOrganization(f.SourceSpace, orgGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", allWarnings, err
		}
		spaceGUID, spaceName = space.GUID, f.SourceSpace
		scope, sourceGUID = "space", spaceGUID
	}

	// App level: resolved within spaceGUID (targeted, --source-space, or both).
	if f.SourceApp != "" {
		app, warnings, err := actor.GetApplicationByNameAndSpace(f.SourceApp, spaceGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			if _, ok := err.(actionerror.ApplicationNotFoundError); ok && f.SourceSpace == "" {
				return "", "", allWarnings, fmt.Errorf(
					"App '%s' not found in space '%s' / org '%s'.\nTIP: If the app is in a different space or org, use --source-space and/or --source-org flags.",
					f.SourceApp,
					config.TargetedSpace().Name,
					config.TargetedOrganization().Name,
				)
			}
			return "", "", allWarnings, err
		}
		scope, sourceGUID = "app", app.GUID
	}

	if scope == "" {
		return "", "", allWarnings, fmt.Errorf("no source specified")
	}

	// Human-readable scope description; annotate cross-space/cross-org modifiers when present.
	var scopeDisplay string
	switch scope {
	case "app":
		scopeDisplay = fmt.Sprintf("scope: app, source: %s", f.SourceApp)
		if f.SourceSpace != "" {
			scopeDisplay += fmt.Sprintf(" (space: %s", spaceName)
			if f.SourceOrg != "" {
				scopeDisplay += fmt.Sprintf(", org: %s", orgName)
			}
			scopeDisplay += ")"
		}
	case "space":
		scopeDisplay = fmt.Sprintf("scope: space, source: %s", f.SourceSpace)
		if f.SourceOrg != "" {
			scopeDisplay += fmt.Sprintf(" (org: %s)", orgName)
		}
	default: // "org"
		scopeDisplay = fmt.Sprintf("scope: org, source: %s", f.SourceOrg)
	}

	return fmt.Sprintf("cf:%s:%s", scope, sourceGUID), scopeDisplay, allWarnings, nil
}
