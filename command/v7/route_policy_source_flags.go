package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
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

	// --source-app (with optional --source-space / --source-org for cross-space/org lookup)
	if f.SourceApp != "" {
		spaceGUID := config.TargetedSpace().GUID
		spaceName := config.TargetedSpace().Name
		orgName := config.TargetedOrganization().Name

		if f.SourceSpace != "" {
			resolvedOrgGUID, resolvedOrgName, orgWarnings, err := resolveOrgGUID(f, actor, config)
			allWarnings = append(allWarnings, orgWarnings...)
			if err != nil {
				return "", "", allWarnings, err
			}
			orgName = resolvedOrgName

			resolvedSpaceGUID, spaceWarnings, err := resolveSpaceGUID(f.SourceSpace, resolvedOrgGUID, actor)
			allWarnings = append(allWarnings, spaceWarnings...)
			if err != nil {
				return "", "", allWarnings, err
			}
			spaceGUID = resolvedSpaceGUID
			spaceName = f.SourceSpace
		}

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

		scopeDisplay := fmt.Sprintf("scope: app, source: %s", f.SourceApp)
		if f.SourceSpace != "" {
			scopeDisplay += fmt.Sprintf(" (space: %s", spaceName)
			if f.SourceOrg != "" {
				scopeDisplay += fmt.Sprintf(", org: %s", orgName)
			}
			scopeDisplay += ")"
		}

		return fmt.Sprintf("cf:app:%s", app.GUID), scopeDisplay, allWarnings, nil
	}

	// --source-space (primary: space-level policy)
	if f.SourceSpace != "" {
		orgGUID, orgName, orgWarnings, err := resolveOrgGUID(f, actor, config)
		allWarnings = append(allWarnings, orgWarnings...)
		if err != nil {
			return "", "", allWarnings, err
		}

		spaceGUID, spaceWarnings, err := resolveSpaceGUID(f.SourceSpace, orgGUID, actor)
		allWarnings = append(allWarnings, spaceWarnings...)
		if err != nil {
			return "", "", allWarnings, err
		}

		scopeDisplay := fmt.Sprintf("scope: space, source: %s", f.SourceSpace)
		if f.SourceOrg != "" {
			scopeDisplay += fmt.Sprintf(" (org: %s)", orgName)
		}

		return fmt.Sprintf("cf:space:%s", spaceGUID), scopeDisplay, allWarnings, nil
	}

	// --source-org (primary: org-level policy)
	if f.SourceOrg != "" {
		org, warnings, err := actor.GetOrganizationByName(f.SourceOrg)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", "", allWarnings, err
		}

		return fmt.Sprintf("cf:org:%s", org.GUID),
			fmt.Sprintf("scope: org, source: %s", f.SourceOrg),
			allWarnings, nil
	}

	return "", "", allWarnings, fmt.Errorf("no source specified")
}

// resolveOrgGUID returns the GUID and name of the org identified by --source-org,
// falling back to the targeted org if --source-org is not set.
func resolveOrgGUID(f RoutePolicySourceFlags, actor Actor, config command.Config) (string, string, v7action.Warnings, error) {
	if f.SourceOrg == "" {
		return config.TargetedOrganization().GUID, config.TargetedOrganization().Name, nil, nil
	}
	org, warnings, err := actor.GetOrganizationByName(f.SourceOrg)
	if err != nil {
		return "", "", v7action.Warnings(warnings), err
	}
	return org.GUID, f.SourceOrg, v7action.Warnings(warnings), nil
}

// resolveSpaceGUID returns the GUID of the named space within the given org.
func resolveSpaceGUID(spaceName, orgGUID string, actor Actor) (string, v7action.Warnings, error) {
	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	if err != nil {
		return "", v7action.Warnings(warnings), err
	}
	return space.GUID, v7action.Warnings(warnings), nil
}
