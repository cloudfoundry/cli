package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . UnmapRouteActor

type UnmapRouteActor interface {
}

type UnmapRouteCommand struct {
	RequiredArgs    flag.AppDomain `positional-args:"yes"`
	Hostname        string         `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path            string         `long:"path" description:"Path used to identify the HTTP route"`
	usage           interface{}    `usage:"CF_NAME unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME unmap-route my-app example.com                              # example.com\n   CF_NAME unmap-route my-app example.com --hostname myhost            # myhost.example.com\n   CF_NAME unmap-route my-app example.com --hostname myhost --path foo # myhost.example.com/foo"`
	relatedCommands interface{}    `related_commands:"delete-route, map-route, routes"`

	UI          command.UI
	Config      command.Config
	Actor       UnmapRouteActor
	SharedActor command.SharedActor
}

func (cmd *UnmapRouteCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd UnmapRouteCommand) Execute(args []string) error {
	return nil
}
