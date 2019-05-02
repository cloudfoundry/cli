package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . DeletePrivateDomainActor

type DeletePrivateDomainActor interface {
	DeletePrivateDomain(domainName string) (v7action.Warnings, error)
	CheckSharedDomain(domainName string) (v7action.Warnings, error)
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
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd DeletePrivateDomainCommand) Execute(args []string) error {
	domain := cmd.RequiredArgs.Domain
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	currentUser, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	shareCheckWarnings, shareCheckErr := cmd.Actor.CheckSharedDomain(domain)

	if shareCheckErr != nil {
		cmd.UI.DisplayWarnings(shareCheckWarnings)
		return shareCheckErr
	}

	cmd.UI.DisplayText("Deleting the private domain will remove associated routes which will make apps with this domain unreachable.")

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the private domain {{.DomainName}}?", map[string]interface{}{
			"DomainName": domain,
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("'{{.DomainName}}' has not been deleted.", map[string]interface{}{
				"DomainName": domain,
			})
			return nil
		}
	}
	cmd.UI.DisplayTextWithFlavor("Deleting private domain {{.DomainName}} as {{.Username}}...", map[string]interface{}{
		"DomainName": domain,
		"Username":   currentUser.Name,
	})

	warnings, err := cmd.Actor.DeletePrivateDomain(domain)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case actionerror.DomainNotFoundError:
			cmd.UI.DisplayTextWithFlavor("Domain {{.DomainName}} not found", map[string]interface{}{
				"DomainName": domain,
			})

		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	cmd.UI.DisplayText("TIP: Run 'cf domains' to view available domains.")

	return nil
}
