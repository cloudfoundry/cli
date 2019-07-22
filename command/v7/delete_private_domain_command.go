package v7

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . DeletePrivateDomainActor

type DeletePrivateDomainActor interface {
	DeleteDomain(domain v7action.Domain) (v7action.Warnings, error)
	GetDomainByName(domainName string) (v7action.Domain, v7action.Warnings, error)
}

type DeletePrivateDomainCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-private-domain DOMAIN [-f]"`
	relatedCommands interface{} `related_commands:"delete-shared-domain, domains, unshare-private-domain"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       DeletePrivateDomainActor
}

func (cmd *DeletePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil, clock.NewClock())

	return nil
}

func (cmd DeletePrivateDomainCommand) Execute(args []string) error {
	domainName := cmd.RequiredArgs.Domain
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain, warnings, err := cmd.Actor.GetDomainByName(domainName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.DomainNotFoundError); ok {
			cmd.UI.DisplayTextWithFlavor("Domain {{.DomainName}} does not exist", map[string]interface{}{
				"DomainName": cmd.RequiredArgs.Domain,
			})
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	if domain.Shared() {
		return fmt.Errorf("Domain '%s' is a shared domain, not a private domain.", domainName)
	}

	cmd.UI.DisplayText("Deleting the private domain will remove associated routes which will make apps with this domain unreachable.")

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the private domain {{.DomainName}}?", map[string]interface{}{
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
	cmd.UI.DisplayTextWithFlavor("Deleting private domain {{.DomainName}} as {{.Username}}...", map[string]interface{}{
		"DomainName": domainName,
		"Username":   currentUser.Name,
	})

	warnings, err = cmd.Actor.DeleteDomain(domain)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	cmd.UI.DisplayText("TIP: Run 'cf domains' to view available domains.")

	return nil
}
