package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreatePrivateDomainActor

type CreatePrivateDomainActor interface {
	CreatePrivateDomain(domainName string, orgName string) (v7action.Warnings, error)
}

type CreatePrivateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME create-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"create-shared-domain, domains, share-private-domain"`

	UI          command.UI
	Config      command.Config
	Actor       CreatePrivateDomainActor
	SharedActor command.SharedActor
}

func (cmd *CreatePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
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
