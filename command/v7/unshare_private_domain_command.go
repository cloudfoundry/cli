package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type UnsharePrivateDomainCommand struct {
	command.BaseCommand

	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME unshare-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"delete-private-domain, domains"`
}

func (cmd UnsharePrivateDomainCommand) Execute(args []string) error {
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

	cmd.UI.DisplayTextWithFlavor("Warning: org {{.Organization}} will no longer be able to access private domain {{.Domain}}",
		map[string]interface{}{
			"Domain":       domain,
			"Organization": orgName,
		})

	response, err := cmd.UI.DisplayBoolPrompt(false,
		"Really unshare private domain {{.Domain}}?",
		map[string]interface{}{
			"Domain": domain,
		},
	)

	if err != nil {
		return err
	}

	if !response {
		cmd.UI.DisplayTextWithFlavor("Domain {{.Domain}} has not been unshared from organization {{.Organization}}",
			map[string]interface{}{
				"Domain":       domain,
				"Organization": orgName,
			})
		return nil
	}

	cmd.UI.DisplayTextWithFlavor("Unsharing domain {{.Domain}} from org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"Domain":       domain,
			"User":         user.Name,
			"Organization": orgName,
		})

	warnings, err := cmd.Actor.UnsharePrivateDomain(domain, orgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
