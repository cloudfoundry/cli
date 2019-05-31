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
	usage           interface{} `usage:"Delete an HTTP route:\n      CF_NAME delete-route DOMAIN [--hostname HOSTNAME] [--path PATH] [-f]\n\n   Delete a TCP route:\n      CF_NAME delete-route DOMAIN [-f]\n\nEXAMPLES:\n   CF_NAME delete-route example.com                              # example.com\n   CF_NAME delete-route example.com --hostname myhost            # myhost.example.com\n   CF_NAME delete-route example.com --hostname myhost --path foo # myhost.example.com/foo"`
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

	cmd.UI.DisplayTextWithFlavor("Deleting route {{.FQDN}}...",
		map[string]interface{}{
			"FQDN": fqdn,
		})

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

//func (cmd DeletePrivateDomainCommand) Execute(args []string) error {

//
//	shareCheckWarnings, shareCheckErr := cmd.Actor.CheckSharedDomain(domain)
//
//	if shareCheckErr != nil {
//		if _, ok := shareCheckErr.(actionerror.DomainNotFoundError); ok {
//			cmd.UI.DisplayTextWithFlavor("Domain {{.DomainName}} does not exist", map[string]interface{}{
//				"DomainName": cmd.RequiredArgs.Domain,
//			})
//			cmd.UI.DisplayOK()
//			return nil
//		}
//
//		cmd.UI.DisplayWarnings(shareCheckWarnings)
//		return shareCheckErr
//	}
//
//	cmd.UI.DisplayText("Deleting the private domain will remove associated routes which will make apps with this domain unreachable.")
//
//	if !cmd.Force {
//		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the private domain {{.DomainName}}?", map[string]interface{}{
//			"DomainName": domain,
//		})
//
//		if promptErr != nil {
//			return promptErr
//		}
//
//		if !response {
//			cmd.UI.DisplayText("'{{.DomainName}}' has not been deleted.", map[string]interface{}{
//				"DomainName": domain,
//			})
//			return nil
//		}
//	}
//	cmd.UI.DisplayTextWithFlavor("Deleting private domain {{.DomainName}} as {{.Username}}...", map[string]interface{}{
//		"DomainName": domain,
//		"Username":   currentUser.Name,
//	})
//
//	warnings, err := cmd.Actor.DeletePrivateDomain(domain)
//	cmd.UI.DisplayWarnings(warnings)
//	if err != nil {
//		switch err.(type) {
//		case actionerror.DomainNotFoundError:
//			cmd.UI.DisplayTextWithFlavor("Domain {{.DomainName}} not found", map[string]interface{}{
//				"DomainName": domain,
//			})
//
//		default:
//			return err
//		}
//	}
//
//	cmd.UI.DisplayOK()
//
//	cmd.UI.DisplayText("TIP: Run 'cf domains' to view available domains.")
//
//	return nil
//}
