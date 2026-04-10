package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/command/flag"
)

type CreateSharedDomainCommand struct {
	BaseCommand

	RequiredArgs       flag.Domain `positional-args:"yes"`
	RouterGroup        string      `long:"router-group" description:"Routes for this domain will use routers in the specified router group"`
	Internal           bool        `long:"internal" description:"Applications that use internal routes communicate directly on the container network"`
	EnforceAccessRules bool        `long:"enforce-access-rules" description:"Enable platform-enforced access control for routes on this domain (requires mTLS domain configuration in GoRouter)"`
	Scope              string      `long:"scope" description:"Operator-level scope boundary for access rules: 'any', 'org', or 'space' (only valid with --enforce-access-rules)"`
	usage              interface{} `usage:"CF_NAME create-shared-domain DOMAIN [--router-group ROUTER_GROUP_NAME | --internal] [--enforce-access-rules [--scope (any|org|space)]]"`
	relatedCommands    interface{} `related_commands:"create-private-domain, domains, add-access-rule, access-rules"`
}

func (cmd CreateSharedDomainCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain

	// Validate that --scope is only used with --enforce-access-rules
	if cmd.Scope != "" && !cmd.EnforceAccessRules {
		return fmt.Errorf("--scope can only be used with --enforce-access-rules")
	}

	// Validate scope values
	if cmd.Scope != "" && cmd.Scope != "any" && cmd.Scope != "org" && cmd.Scope != "space" {
		return fmt.Errorf("--scope must be one of: any, org, space")
	}

	cmd.UI.DisplayTextWithFlavor("Creating shared domain {{.Domain}} as {{.User}}...",
		map[string]interface{}{
			"Domain": domain,
			"User":   user.Name,
		})

	warnings, err := cmd.Actor.CreateSharedDomain(domain, cmd.Internal, cmd.RouterGroup, cmd.EnforceAccessRules, cmd.Scope)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	
	if cmd.EnforceAccessRules {
		cmd.UI.DisplayText("TIP: Domain '{{.Domain}}' is a shared identity-aware domain with access rule enforcement enabled. Routes on this domain require access rules to allow traffic.",
			map[string]interface{}{
				"Domain": domain,
			})
	} else {
		cmd.UI.DisplayText("TIP: Domain '{{.Domain}}' is shared with all orgs. Run 'cf domains' to view available domains.",
			map[string]interface{}{
				"Domain": domain,
			})
	}
	return nil
}
