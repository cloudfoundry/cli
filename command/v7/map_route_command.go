package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . MapRouteActor

type MapRouteActor interface{}

type MapRouteCommand struct {
	RequiredArgs    flag.AppDomain `positional-args:"yes"`
	usage           interface{}    `usage:"CF_NAME map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME map-route my-app example.com                              # example.com\n   CF_NAME map-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME map-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo"`
	Hostname        string         `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            string         `long:"path" description:"Path for the HTTP route"`
	relatedCommands interface{}    `related_commands:"create-route, routes, unmap-route"`

	UI          command.UI
	Config      command.Config
	Actor       MapRouteActor
	SharedActor command.SharedActor
}

func (cmd *MapRouteCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd MapRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	return nil
}
