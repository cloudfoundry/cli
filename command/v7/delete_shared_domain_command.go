package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
)

type DeleteSharedDomainCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain `positional-args:"yes"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-shared-domain DOMAIN [-f]"`
	relatedCommands interface{} `related_commands:"delete-private-domain, domains"`
}

func (cmd DeleteSharedDomainCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}
	domainName := cmd.RequiredArgs.Domain

	currentUser, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}
	cmd.UI.DisplayText("This action impacts all orgs using this domain.")
	cmd.UI.DisplayText("Deleting the domain will remove associated routes which will make apps with this domain, in any org, unreachable.")

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the shared domain {{.DomainName}}?", map[string]interface{}{
			"DomainName": domainName,
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("'{{.DomainName}}' has not been deleted.", map[string]interface{}{
				"DomainName": domainName,
			})
			return nil
		}
	}
	cmd.UI.DisplayTextWithFlavor("Deleting domain {{.DomainName}} as {{.Username}}...", map[string]interface{}{
		"DomainName": domainName,
		"Username":   currentUser.Name,
	})

	domain, warnings, err := cmd.Actor.GetDomainByName(domainName)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		if _, ok := err.(actionerror.DomainNotFoundError); ok {
			cmd.UI.DisplayWarning("Domain '{{.DomainName}}' does not exist.", map[string]interface{}{
				"DomainName": cmd.RequiredArgs.Domain,
			})
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	// Private domains always have an organization guid; shared domains, never.
	if domain.OrganizationGUID != "" {
		return translatableerror.NotSharedDomainError{DomainName: cmd.RequiredArgs.Domain}
	}

	warnings, err = cmd.Actor.DeleteDomain(domain)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
