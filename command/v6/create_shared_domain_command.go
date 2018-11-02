package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . CreateSharedDomainActor

type CreateSharedDomainActor interface {
	GetRouterGroupByName(string, v2action.RouterClient) (v2action.RouterGroup, error)
	CreateSharedDomain(string, v2action.RouterGroup) (v2action.Warnings, error)
}

type CreateSharedDomainCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	RouterGroup     string      `long:"router-group" description:"Routes for this domain will be configured only on the specified router group"`
	usage           interface{} `usage:"CF_NAME create-shared-domain DOMAIN [--router-group ROUTER_GROUP]"`
	relatedCommands interface{} `related_commands:"create-domain, domains, router-groups"`

	UI           command.UI
	Config       command.Config
	Actor        CreateSharedDomainActor
	SharedActor  command.SharedActor
	RouterClient v2action.RouterClient
}

func (cmd *CreateSharedDomainCommand) Setup(config command.Config, ui command.UI) error {
	ccClient, uaaClient, err := shared.NewClients(config, ui, true)

	if err != nil {
		return err
	}

	routerClient, err := shared.NewRouterClient(config, ui, uaaClient)

	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.RouterClient = routerClient
	cmd.SharedActor = sharedaction.NewActor(config)
	cmd.Config = config
	cmd.UI = ui
	return nil
}

func (cmd CreateSharedDomainCommand) Execute(args []string) error {
	username, err := cmd.SharedActor.RequireCurrentUser()
	cmd.UI.DisplayTextWithFlavor("Creating shared domain {{.Domain}} as {{.User}}...",
		map[string]interface{}{
			"Domain": cmd.RequiredArgs.Domain,
			"User":   username,
		})

	if err != nil {
		return err
	}

	var routerGroup v2action.RouterGroup

	if cmd.RouterGroup != "" {
		routerGroup, err = cmd.Actor.GetRouterGroupByName(cmd.RouterGroup, cmd.RouterClient)
		if err != nil {
			return err
		}
	}

	warnings, err := cmd.Actor.CreateSharedDomain(cmd.RequiredArgs.Domain, routerGroup)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
