package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . CreateRouteActor

type CreateRouteActor interface {
	CreateRoute(spaceName string, domainName string) (v7action.Warnings, error)
}

type CreateRouteCommand struct {
	RequiredArgs    flag.Domain `positional-args:"yes"`
	usage           interface{} `usage:"CF_NAME create-route DOMAIN [--hostname HOSTNAME] [--path PATH]\n\nEXAMPLES:\n   CF_NAME create-route example.com                              # example.com\n   CF_NAME create-route example.com --hostname myapp            # myapp.example.com\n   CF_NAME create-route example.com --hostname myapp --path foo # myapp.example.com/foo"`
	Hostname        string      `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            string      `long:"path" description:"Path for the HTTP route"`
	relatedCommands interface{} `related_commands:"check-route, domains, map-route, routes, unmap route"`

	UI          command.UI
	Config      command.Config
	Actor       CreateRouteActor
	SharedActor command.SharedActor
}

func (cmd *CreateRouteCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd CreateRouteCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	domain := cmd.RequiredArgs.Domain
	spaceName := cmd.Config.TargetedSpace().Name
	orgName := cmd.Config.TargetedOrganization().Name

	cmd.UI.DisplayTextWithFlavor("Creating route {{.Domain}} for org {{.Organization}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"Domain":       domain,
			"User":         user.Name,
			"Space":        spaceName,
			"Organization": orgName,
		})

	warnings, err := cmd.Actor.CreateRoute(spaceName, domain)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteAlreadyExistsError); ok {
			cmd.UI.DisplayTextWithFlavor("Route {{.DomainName}} already exists.", map[string]interface{}{
				"DomainName": domain,
			})
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	cmd.UI.DisplayText("Route {{.Domain}} has been created.",
		map[string]interface{}{
			"Domain": domain,
		})

	cmd.UI.DisplayOK()
	return nil
}
