package v7

import (
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/sorting"
	"code.cloudfoundry.org/cli/v8/util/ui"
)

type DomainsCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME domains\n\nEXAMPLES:\n   CF_NAME domains\n   CF_NAME domains --labels 'environment in (production,staging),tier in (backend)'\n   CF_NAME domains --labels 'env=dev,!chargeback-code,tier in (backend,worker)'"`
	relatedCommands interface{} `related_commands:"create-private-domain, create-route, create-shared-domain, routes, set-label"`
	Labels          string      `long:"labels" description:"Selector to filter domains by labels"`
}

func (cmd DomainsCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	targetedOrg := cmd.Config.TargetedOrganization()
	cmd.UI.DisplayTextWithFlavor("Getting domains in org {{.CurrentOrg}} as {{.CurrentUser}}...\n", map[string]interface{}{
		"CurrentOrg":  targetedOrg.Name,
		"CurrentUser": currentUser.Name,
	})

	domains, warnings, err := cmd.Actor.GetOrganizationDomains(targetedOrg.GUID, cmd.Labels)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	sort.Slice(domains, func(i, j int) bool { return sorting.LessIgnoreCase(domains[i].Name, domains[j].Name) })

	if len(domains) > 0 {
		cmd.displayDomainsTable(domains)
	} else {
		cmd.UI.DisplayText("No domains found.")
	}
	return nil
}

func (cmd DomainsCommand) displayDomainsTable(domains []resources.Domain) {
	showRoutePoliciesCol := command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionRoutePolicies) == nil

	headers := []string{
		cmd.UI.TranslateText("name"),
		cmd.UI.TranslateText("availability"),
		cmd.UI.TranslateText("internal"),
		cmd.UI.TranslateText("protocols"),
	}
	if showRoutePoliciesCol {
		headers = append(headers, cmd.UI.TranslateText("route policies"))
	}

	domainsTable := [][]string{headers}

	for _, domain := range domains {
		var availability string
		var internal string

		if domain.Shared() {
			availability = cmd.UI.TranslateText("shared")
		} else {
			availability = cmd.UI.TranslateText("private")
		}

		if domain.Internal.IsSet && domain.Internal.Value {
			internal = cmd.UI.TranslateText("true")
		}

		row := []string{
			domain.Name,
			availability,
			internal,
			strings.Join(domain.Protocols, ","),
		}

		if showRoutePoliciesCol {
			row = append(row, routePoliciesDisplay(domain))
		}

		domainsTable = append(domainsTable, row)
	}

	cmd.UI.DisplayTableWithHeader("", domainsTable, ui.DefaultTableSpacePadding)
}

// routePoliciesDisplay returns the combined route policies cell value for a domain:
//   - blank   — enforcement not enabled
//   - "enforced"          — enabled, no scope restriction
//   - "enforced (org)"    — enabled, org-scoped
//   - "enforced (space)"  — enabled, space-scoped
//   - "enforced (any)"    — enabled, any-scoped
func routePoliciesDisplay(d resources.Domain) string {
	if !d.EnforceRoutePolicies.IsSet || !d.EnforceRoutePolicies.Value {
		return ""
	}
	if d.RoutePoliciesScope == "" {
		return "enforced"
	}
	return fmt.Sprintf("enforced (%s)", d.RoutePoliciesScope)
}
