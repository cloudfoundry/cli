package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SharePrivateDomainActor

type SharePrivateDomainActor interface {
	SharePrivateDomain(domainName string, orgName string) (v7action.Warnings, error)
}

type SharePrivateDomainCommand struct {
	RequiredArgs    flag.OrgDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME share-private-domain ORG DOMAIN"`
	relatedCommands interface{}    `related_commands:"create-private-domain, domains, unshare-private-domain"`

	UI          command.UI
	Config      command.Config
	Actor       SharePrivateDomainActor
	SharedActor command.SharedActor
}

func (cmd *SharePrivateDomainCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
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
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
