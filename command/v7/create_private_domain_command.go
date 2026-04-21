package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type CreatePrivateDomainCommand struct {
	BaseCommand

	RequiredArgs         flag.OrgDomain `positional-args:"yes"`
	EnforceRoutePolicies bool           `long:"enforce-route-policies" description:"Enable platform-enforced access control for routes on this domain (requires mTLS domain configuration in GoRouter)"`
	Scope                string         `long:"scope" description:"Operator-level scope boundary for route policies: 'any', 'org', or 'space' (only valid with --enforce-route-policies)"`
	usage                interface{}    `usage:"CF_NAME create-private-domain ORG DOMAIN [--enforce-route-policies [--scope (any|org|space)]]"`
	relatedCommands      interface{}    `related_commands:"create-shared-domain, domains, share-private-domain, add-route-policy, route-policies"`
}

func (cmd CreatePrivateDomainCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	orgName := cmd.RequiredArgs.Organization

	// Validate that --scope is only used with --enforce-route-policies
	if cmd.Scope != "" && !cmd.EnforceRoutePolicies {
		return fmt.Errorf("--scope can only be used with --enforce-route-policies")
	}

	// Validate scope values
	if cmd.Scope != "" && cmd.Scope != "any" && cmd.Scope != "org" && cmd.Scope != "space" {
		return fmt.Errorf("--scope must be one of: any, org, space")
	}

	cmd.UI.DisplayTextWithFlavor("Creating private domain {{.Domain}} for org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"Domain":       domain,
			"User":         user.Name,
			"Organization": orgName,
		})

	warnings, err := cmd.Actor.CreatePrivateDomain(domain, orgName, cmd.EnforceRoutePolicies, cmd.Scope)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if e, ok := err.(ccerror.UnprocessableEntityError); ok {
			inUse := fmt.Sprintf("The domain name \"%s\" is already in use", domain)
			if e.Message == inUse {
				cmd.UI.DisplayWarning(err.Error())
				cmd.UI.DisplayOK()
				return nil
			}
		}
		return err
	}

	cmd.UI.DisplayOK()

	if cmd.EnforceRoutePolicies {
		cmd.UI.DisplayText("TIP: Domain '{{.Domain}}' is a private identity-aware domain with route policy enforcement enabled. Routes on this domain require route policies to allow traffic.",
			map[string]interface{}{
				"Domain": domain,
			})
	} else {
		cmd.UI.DisplayText("TIP: Domain '{{.Domain}}' is a private domain. Run 'cf share-private-domain' to share this domain with a different org.",
			map[string]interface{}{
				"Domain": domain,
			})
	}
	return nil
}
