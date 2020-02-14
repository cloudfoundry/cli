package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type CreatePrivateDomainCommand struct {
	BaseCommand

	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME create-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"create-shared-domain, domains, share-private-domain"`
}

func (cmd CreatePrivateDomainCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	orgName := cmd.RequiredArgs.Organization

	cmd.UI.DisplayTextWithFlavor("Creating private domain {{.Domain}} for org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"Domain":       domain,
			"User":         user.Name,
			"Organization": orgName,
		})

	warnings, err := cmd.Actor.CreatePrivateDomain(domain, orgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	cmd.UI.DisplayText("TIP: Domain '{{.Domain}}' is a private domain. Run 'cf share-private-domain' to share this domain with a different org.",
		map[string]interface{}{
			"Domain": domain,
		})
	return nil
}
