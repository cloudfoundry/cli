package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type CreateSharedDomainCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain `positional-args:"yes"`
	RouterGroup     string      `long:"router-group" description:"Routes for this domain will use routers in the specified router group"`
	Internal        bool        `long:"internal" description:"Applications that use internal routes communicate directly on the container network"`
	usage           interface{} `usage:"CF_NAME create-shared-domain DOMAIN [--router-group ROUTER_GROUP_NAME | --internal]"`
	relatedCommands interface{} `related_commands:"create-private-domain, domains"`
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

	cmd.UI.DisplayTextWithFlavor("Creating shared domain {{.Domain}} as {{.User}}...",
		map[string]interface{}{
			"Domain": domain,
			"User":   user.Name,
		})

	warnings, err := cmd.Actor.CreateSharedDomain(domain, cmd.Internal, cmd.RouterGroup)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Domain '{{.Domain}}' is shared with all orgs. Run 'cf domains' to view available domains.",
		map[string]interface{}{
			"Domain": domain,
		})
	return nil
}
