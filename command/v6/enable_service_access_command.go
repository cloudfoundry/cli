package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . EnableServiceAccessActor

type EnableServiceAccessActor interface {
	EnablePlanForOrg(serviceName, servicePlanName, orgName string) (v2action.Warnings, error)
	EnableServiceForOrg(serviceName, orgName string) (v2action.Warnings, error)
	EnablePlanForAllOrgs(serviceName, servicePlanName string) (v2action.Warnings, error)
	EnableServiceForAllOrgs(serviceName string) (v2action.Warnings, error)
}

type EnableServiceAccessCommand struct {
	RequiredArgs    flag.Service `positional-args:"yes"`
	Organization    string       `short:"o" description:"Enable access for a specified organization"`
	ServicePlan     string       `short:"p" description:"Enable access to a specified service plan"`
	usage           interface{}  `usage:"CF_NAME enable-service-access SERVICE [-p PLAN] [-o ORG]"`
	relatedCommands interface{}  `related_commands:"marketplace, service-access, service-brokers"`

	UI          command.UI
	SharedActor command.SharedActor
	Actor       EnableServiceAccessActor
	Config      command.Config
}

func (cmd *EnableServiceAccessCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}

	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd EnableServiceAccessCommand) Execute(args []string) error {
	if len(args) > 0 {
		return translatableerror.TooManyArgumentsError{
			ExtraArgument: args[0],
		}
	}

	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	serviceName := cmd.RequiredArgs.Service
	servicePlanName := cmd.ServicePlan
	orgName := cmd.Organization
	var warnings v2action.Warnings

	if servicePlanName != "" && orgName != "" {
		cmd.UI.DisplayTextWithFlavor("Enabling access to plan {{.ServicePlan}} of service {{.Service}} for org {{.Organization}} as {{.User}}...",
			map[string]interface{}{
				"ServicePlan":  servicePlanName,
				"Service":      serviceName,
				"Organization": orgName,
				"User":         user.Name,
			})
		warnings, err = cmd.Actor.EnablePlanForOrg(serviceName, servicePlanName, orgName)
	} else if orgName != "" {
		cmd.UI.DisplayTextWithFlavor("Enabling access to all plans of service {{.Service}} for the org {{.Organization}} as {{.User}}...",
			map[string]interface{}{
				"Service":      serviceName,
				"Organization": orgName,
				"User":         user.Name,
			})
		warnings, err = cmd.Actor.EnableServiceForOrg(serviceName, orgName)
	} else if servicePlanName != "" {
		cmd.UI.DisplayTextWithFlavor("Enabling access of plan {{.ServicePlan}} for service {{.Service}} as {{.User}}...",
			map[string]interface{}{
				"ServicePlan": servicePlanName,
				"Service":     serviceName,
				"User":        user.Name,
			})
		warnings, err = cmd.Actor.EnablePlanForAllOrgs(serviceName, servicePlanName)
	} else {
		cmd.UI.DisplayTextWithFlavor("Enabling access to all plans of service {{.Service}} for all orgs as {{.User}}...",
			map[string]interface{}{
				"Service": serviceName,
				"User":    user.Name,
			})
		warnings, err = cmd.Actor.EnableServiceForAllOrgs(serviceName)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
