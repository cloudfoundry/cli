package serviceaccess

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DisableServiceAccess struct {
	ui             terminal.UI
	config         coreconfig.Reader
	actor          actors.ServicePlanActor
	tokenRefresher authentication.TokenRefresher
}

func init() {
	commandregistry.Register(&DisableServiceAccess{})
}

func (cmd *DisableServiceAccess) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Disable access to a specified service plan")}
	fs["o"] = &flags.StringFlag{ShortName: "o", Usage: T("Disable access for a specified organization")}

	return commandregistry.CommandMetadata{
		Name:        "disable-service-access",
		Description: T("Disable access to a service or service plan for one or all orgs"),
		Usage: []string{
			"CF_NAME disable-service-access SERVICE [-p PLAN] [-o ORG]",
		},
		Flags: fs,
	}
}

func (cmd *DisableServiceAccess) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("disable-service-access"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *DisableServiceAccess) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.actor = deps.ServicePlanHandler
	cmd.tokenRefresher = deps.RepoLocator.GetAuthenticationRepository()
	return cmd
}

func (cmd *DisableServiceAccess) Execute(c flags.FlagContext) error {
	_, err := cmd.tokenRefresher.RefreshAuthToken()
	if err != nil {
		return err
	}

	serviceName := c.Args()[0]
	planName := c.String("p")
	orgName := c.String("o")

	if planName != "" && orgName != "" {
		err = cmd.disablePlanAndOrgForService(serviceName, planName, orgName)
	} else if planName != "" {
		err = cmd.disableSinglePlanForService(serviceName, planName)
	} else if orgName != "" {
		err = cmd.disablePlansForSingleOrgForService(serviceName, orgName)
	} else {
		err = cmd.disableServiceForAll(serviceName)
	}
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}

func (cmd *DisableServiceAccess) disableServiceForAll(serviceName string) error {
	cmd.ui.Say(T("Disabling access to all plans of service {{.ServiceName}} for all orgs as {{.UserName}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceName),
			"UserName":    terminal.EntityNameColor(cmd.config.Username()),
		},
	))

	return cmd.actor.UpdateAllPlansForService(serviceName, false)
}

func (cmd *DisableServiceAccess) disablePlanAndOrgForService(serviceName string, planName string, orgName string) error {
	cmd.ui.Say(T("Disabling access to plan {{.PlanName}} of service {{.ServiceName}} for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"PlanName":    terminal.EntityNameColor(planName),
			"ServiceName": terminal.EntityNameColor(serviceName),
			"OrgName":     terminal.EntityNameColor(orgName),
			"Username":    terminal.EntityNameColor(cmd.config.Username()),
		},
	))

	return cmd.actor.UpdatePlanAndOrgForService(serviceName, planName, orgName, false)
}

func (cmd *DisableServiceAccess) disableSinglePlanForService(serviceName string, planName string) error {
	cmd.ui.Say(T("Disabling access of plan {{.PlanName}} for service {{.ServiceName}} as {{.Username}}...",
		map[string]interface{}{
			"PlanName":    terminal.EntityNameColor(planName),
			"ServiceName": terminal.EntityNameColor(serviceName),
			"Username":    terminal.EntityNameColor(cmd.config.Username()),
		},
	))

	return cmd.actor.UpdateSinglePlanForService(serviceName, planName, false)
}

func (cmd *DisableServiceAccess) disablePlansForSingleOrgForService(serviceName string, orgName string) error {
	cmd.ui.Say(T("Disabling access to all plans of service {{.ServiceName}} for the org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceName),
			"OrgName":     terminal.EntityNameColor(orgName),
			"Username":    terminal.EntityNameColor(cmd.config.Username()),
		},
	))

	return cmd.actor.UpdateOrgForService(serviceName, orgName, false)
}
