package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . CreateSharedDomainActor

type CreateSharedDomainActor interface {
	CreateSharedDomain(domainName string, internal bool) (v7action.Warnings, error)
}

type CreateSharedDomainCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	Internal        bool        `long:"internal" description:"Applications that use internal routes communicate directly on the container network"`
	usage           interface{} `usage:"CF_NAME create-shared-domain DOMAIN [--internal]"`
	relatedCommands interface{} `related_commands:"create-private-domain, domains"`

	UI          command.UI
	Config      command.Config
	Actor       CreateSharedDomainActor
	SharedActor command.SharedActor
}

func (cmd *CreateSharedDomainCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)
	return nil
}

func (cmd CreateSharedDomainCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain

	cmd.UI.DisplayTextWithFlavor("Creating shared domain {{.Domain}} as {{.User}}...",
		map[string]interface{}{
			"Domain": domain,
			"User":   user.Name,
		})

	warnings, err := cmd.Actor.CreateSharedDomain(domain, cmd.Internal)
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
