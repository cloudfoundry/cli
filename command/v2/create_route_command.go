package v2

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . CreateRouteActor

type CreateRouteActor interface {
	CloudControllerAPIVersion() string
	CreateRouteWithExistenceCheck(orgGUID string, spaceName string, route v2action.Route, generatePort bool) (v2action.Route, v2action.Warnings, error)
}

type CreateRouteCommand struct {
	RequiredArgs    flag.SpaceDomain `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            string           `long:"path" description:"Path for the HTTP route"`
	Port            flag.Port        `long:"port" description:"Port for the TCP route"`
	RandomPort      bool             `long:"random-port" description:"Create a random port for the TCP route"`
	usage           interface{}      `usage:"Create an HTTP route:\n      CF_NAME create-route SPACE DOMAIN [--hostname HOSTNAME] [--path PATH]\n\n   Create a TCP route:\n      CF_NAME create-route SPACE DOMAIN (--port PORT | --random-port)\n\nEXAMPLES:\n   CF_NAME create-route my-space example.com                             # example.com\n   CF_NAME create-route my-space example.com --hostname myapp            # myapp.example.com\n   CF_NAME create-route my-space example.com --hostname myapp --path foo # myapp.example.com/foo\n   CF_NAME create-route my-space example.com --port 5000                 # example.com:5000"`
	relatedCommands interface{}      `related_commands:"check-route, domains, map-route"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateRouteActor
}

func (cmd *CreateRouteCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd CreateRouteCommand) Execute(args []string) error {
	err := cmd.validateArguments()
	if err != nil {
		return err
	}

	err = cmd.minimumFlagVersions()
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	route := v2action.Route{
		Domain: v2action.Domain{Name: cmd.RequiredArgs.Domain},
		Host:   cmd.Hostname,
		Path:   cmd.Path,
		Port:   cmd.Port.NullInt,
	}

	cmd.UI.DisplayTextWithFlavor("Creating route {{.Route}} for org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"Route":     route,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.RequiredArgs.Space,
		"Username":  user.Name,
	})

	createdRoute, warnings, err := cmd.Actor.CreateRouteWithExistenceCheck(cmd.Config.TargetedOrganization().GUID, cmd.RequiredArgs.Space, route, cmd.RandomPort)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.RouteAlreadyExistsError); ok {
			cmd.UI.DisplayWarning("Route {{.Route}} already exists.", map[string]interface{}{
				"Route": route,
			})
			cmd.UI.DisplayOK()
			return nil
		}

		return err
	}

	cmd.UI.DisplayTextWithFlavor("Route {{.Route}} has been created.", map[string]interface{}{
		"Route": createdRoute,
	})

	cmd.UI.DisplayOK()

	return nil
}

func (cmd CreateRouteCommand) minimumFlagVersions() error {
	ccVersion := cmd.Actor.CloudControllerAPIVersion()
	if err := command.MinimumAPIVersionCheck(ccVersion, ccversion.MinVersionHTTPRoutePath, "Option '--path'"); cmd.Path != "" && err != nil {
		return err
	}
	if err := command.MinimumAPIVersionCheck(ccVersion, ccversion.MinVersionTCPRouting, "Option '--port'"); cmd.Port.IsSet && err != nil {
		return err
	}
	if err := command.MinimumAPIVersionCheck(ccVersion, ccversion.MinVersionTCPRouting, "Option '--random-port'"); cmd.RandomPort && err != nil {
		return err
	}
	return nil
}

func (cmd CreateRouteCommand) validateArguments() error {
	var failedArgs []string

	if cmd.Hostname != "" {
		failedArgs = append(failedArgs, "--hostname")
	}
	if cmd.Path != "" {
		failedArgs = append(failedArgs, "--path")
	}
	if cmd.Port.IsSet {
		failedArgs = append(failedArgs, "--port")
	}
	if cmd.RandomPort {
		failedArgs = append(failedArgs, "--random-port")
	}

	switch {
	case (cmd.Hostname != "" || cmd.Path != "") && (cmd.Port.IsSet || cmd.RandomPort),
		cmd.Port.IsSet && cmd.RandomPort:
		return translatableerror.ArgumentCombinationError{Args: failedArgs}
	}

	return nil
}
