package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . DeleteRouteActor

type DeleteRouteActor interface {
	DeleteRoute(domainName, hostname, path string) (v7action.Warnings, error)
}

type DeleteRouteCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME delete-route DOMAIN [--hostname HOSTNAME] [--path PATH] [-f]"`
	Force           bool        `short:"f" description:"Force deletion without confirmation"`
	Hostname        string      `long:"hostname" short:"n" description:"Hostname used to identify the HTTP route (required for shared domains)"`
	Path            string      `long:"path" description:"Path used to identify the HTTP route"`
	relatedCommands interface{} `related_commands:"delete-orphaned-routes, routes, unmap-route"`

	UI          command.UI
	Config      command.Config
	Actor       DeleteRouteActor
	SharedActor command.SharedActor
}

func (cmd *DeleteRouteCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd DeleteRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	hostname := cmd.Hostname
	pathName := cmd.Path
	fqdn := desiredFQDN(domain, hostname, pathName)

	cmd.UI.DisplayText("This action impacts all apps using this route.")
	cmd.UI.DisplayText("Deleting the route will remove associated apps which will make apps with this route unreachable.")

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the route {{.FQDN}}?", map[string]interface{}{
			"FQDN": fqdn,
		})

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("'{{.FQDN}}' has not been deleted.", map[string]interface{}{
				"FQDN": fqdn,
			})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting route {{.FQDN}}...",
		map[string]interface{}{
			"FQDN": fqdn,
		})

	warnings, err := cmd.Actor.DeleteRoute(domain, hostname, pathName)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteNotFoundError); ok {
			cmd.UI.DisplayText(`Unable to delete. ` + err.Error())
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
