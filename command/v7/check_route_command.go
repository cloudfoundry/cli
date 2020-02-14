package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CheckRouteActor

type CheckRouteActor interface {
	CheckRoute(domainName string, hostname string, path string) (bool, v7action.Warnings, error)
}

type CheckRouteCommand struct {
	BaseCommand

	RequiredArgs    flag.Domain      `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route"`
	Path            flag.V7RoutePath `long:"path" description:"Path for the route"`
	usage           interface{}      `usage:"CF_NAME check-route DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME check-route example.com                      # example.com\n   CF_NAME check-route example.com -n myhost --path foo # myhost.example.com/foo\n   CF_NAME check-route example.com --path foo           # example.com/foo"`
	relatedCommands interface{}      `related_commands:"create-route, delete-route, routes"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CheckRouteActor
}

func (cmd *CheckRouteCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd CheckRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Checking for route...")

	path := cmd.Path.Path
	matches, warnings, err := cmd.Actor.CheckRoute(cmd.RequiredArgs.Domain, cmd.Hostname, path)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return err
	}

	formatParams := map[string]interface{}{
		"FQDN": desiredFQDN(cmd.RequiredArgs.Domain, cmd.Hostname, path),
	}

	if matches {
		cmd.UI.DisplayText("Route '{{.FQDN}}' does exist.", formatParams)
	} else {
		cmd.UI.DisplayText("Route '{{.FQDN}}' does not exist.", formatParams)
	}

	cmd.UI.DisplayOK()

	return nil
}
