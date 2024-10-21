package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
)

type UnbindRouteServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.RouteServiceArgs `positional-args:"yes"`
	Force           bool                  `short:"f" description:"Force unbinding without confirmation"`
	Hostname        string                `long:"hostname" short:"n" description:"Hostname used in combination with DOMAIN to specify the route to unbind"`
	Path            flag.V7RoutePath      `long:"path" description:"Path used in combination with HOSTNAME and DOMAIN to specify the route to unbind"`
	Wait            bool                  `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands interface{}           `related_commands:"delete-service, routes, services"`
}

func (cmd UnbindRouteServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if !cmd.Force {
		delete, err := cmd.displayPrompt()
		if err != nil {
			return err
		}

		if !delete {
			cmd.UI.DisplayText("Unbind cancelled")
			return nil
		}
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.DeleteRouteBinding(v7action.DeleteRouteBindingParams{
		SpaceGUID:           cmd.Config.TargetedSpace().GUID,
		ServiceInstanceName: cmd.RequiredArgs.ServiceInstance,
		DomainName:          cmd.RequiredArgs.Domain,
		Hostname:            cmd.Hostname,
		Path:                cmd.Path.Path,
	})
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
	case actionerror.RouteBindingNotFoundError:
		cmd.displayNotFound()
		return nil
	default:
		return err
	}

	completed, err := shared.WaitForResult(stream, cmd.UI, cmd.Wait)
	switch {
	case err != nil:
		return err
	case completed:
		cmd.UI.DisplayOK()
		return nil
	default:
		cmd.UI.DisplayOK()
		cmd.UI.DisplayText("Unbinding in progress.")
		return nil
	}
}

func (UnbindRouteServiceCommand) Usage() string {
	return `CF_NAME unbind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE [-f]`
}

func (UnbindRouteServiceCommand) Examples() string {
	return `CF_NAME unbind-route-service example.com --hostname myapp --path foo myratelimiter`
}

func (cmd UnbindRouteServiceCommand) displayPrompt() (bool, error) {
	delete, err := cmd.UI.DisplayBoolPrompt(
		false,
		"Really unbind route {{.URL}} from service instance {{.ServiceInstance}}?",
		map[string]interface{}{
			"URL":             desiredURL(cmd.RequiredArgs.Domain, cmd.Hostname, cmd.Path.Path, 0),
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
		},
	)
	if err != nil {
		return false, err
	}

	return delete, nil
}

func (cmd UnbindRouteServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Unbinding route {{.URL}} from service instance {{.ServiceInstance}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"URL":             desiredURL(cmd.RequiredArgs.Domain, cmd.Hostname, cmd.Path.Path, 0),
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
			"User":            user.Name,
			"Space":           cmd.Config.TargetedSpace().Name,
			"Org":             cmd.Config.TargetedOrganization().Name,
		},
	)

	return nil
}

func (cmd UnbindRouteServiceCommand) displayNotFound() {
	cmd.UI.DisplayText(
		"Route {{.URL}} was not bound to service instance {{.ServiceInstance}}.",
		map[string]interface{}{
			"URL":             desiredURL(cmd.RequiredArgs.Domain, cmd.Hostname, cmd.Path.Path, 0),
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
		},
	)
	cmd.UI.DisplayOK()
}
