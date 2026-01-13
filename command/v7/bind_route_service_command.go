package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
	"code.cloudfoundry.org/cli/v9/types"
)

type BindRouteServiceCommand struct {
	BaseCommand

	RequiredArgs    flag.RouteServiceArgs         `positional-args:"yes"`
	Parameters      flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided inline or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	Hostname        string                        `long:"hostname" short:"n" description:"Hostname used in combination with DOMAIN to specify the route to bind"`
	Path            flag.V7RoutePath              `long:"path" description:"Path used in combination with HOSTNAME and DOMAIN to specify the route to bind"`
	Wait            bool                          `short:"w" long:"wait" description:"Wait for the operation to complete"`
	relatedCommands interface{}                   `related_commands:"routes, services"`
}

func (cmd BindRouteServiceCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}

	if err := cmd.displayIntro(); err != nil {
		return err
	}

	stream, warnings, err := cmd.Actor.CreateRouteBinding(v7action.CreateRouteBindingParams{
		SpaceGUID:           cmd.Config.TargetedSpace().GUID,
		ServiceInstanceName: cmd.RequiredArgs.ServiceInstance,
		DomainName:          cmd.RequiredArgs.Domain,
		Hostname:            cmd.Hostname,
		Path:                cmd.Path.Path,
		Parameters:          types.OptionalObject(cmd.Parameters),
	})
	cmd.UI.DisplayWarnings(warnings)
	switch err.(type) {
	case nil:
	case actionerror.ResourceAlreadyExistsError:
		cmd.displayAlreadyExists()
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
		cmd.UI.DisplayText("Binding in progress.")
		return nil
	}
}

func (cmd BindRouteServiceCommand) Usage() string {
	return `CF_NAME bind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]`
}

func (cmd BindRouteServiceCommand) Examples() string {
	return `
CF_NAME bind-route-service example.com --hostname myapp --path foo myratelimiter
CF_NAME bind-route-service example.com myratelimiter -c file.json
CF_NAME bind-route-service example.com myratelimiter -c '{"valid":"json"}'

In Windows PowerShell use double-quoted, escaped JSON: "{\"valid\":\"json\"}"
In Windows Command Line use single-quoted, escaped JSON: '{\"valid\":\"json\"}'
`
}

func (cmd BindRouteServiceCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Binding route {{.URL}} to service instance {{.ServiceInstance}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
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

func (cmd BindRouteServiceCommand) displayAlreadyExists() {
	cmd.UI.DisplayText(
		"Route {{.URL}} is already bound to service instance {{.ServiceInstance}}.",
		map[string]interface{}{
			"URL":             desiredURL(cmd.RequiredArgs.Domain, cmd.Hostname, cmd.Path.Path, 0),
			"ServiceInstance": cmd.RequiredArgs.ServiceInstance,
		},
	)
	cmd.UI.DisplayOK()
}
