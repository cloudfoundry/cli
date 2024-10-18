package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/command/flag"
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

	user, err := cmd.Actor.GetCurrentUser()
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

	cmd.UI.DisplayText("TIP: Domain '{{.Domain}}' is a private domain. Run 'cf share-private-domain' to share this domain with a different org.",
		map[string]interface{}{
			"Domain": domain,
		})
	return nil
}
