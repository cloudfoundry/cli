package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type SharePrivateDomainCommand struct {
	command.BaseCommand

	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME share-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"create-private-domain, domains, unshare-private-domain"`
}

func (cmd SharePrivateDomainCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Sharing domain {{.Domain}} with org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"Domain":       domain,
			"User":         user.Name,
			"Organization": orgName,
		})

	warnings, err := cmd.Actor.SharePrivateDomain(domain, orgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
