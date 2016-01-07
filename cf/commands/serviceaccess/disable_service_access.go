package serviceaccess

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type DisableServiceAccess struct {
	ui             terminal.UI
	config         core_config.Reader
	actor          actors.ServicePlanActor
	tokenRefresher authentication.TokenRefresher
}

func init() {
	command_registry.Register(&DisableServiceAccess{})
}

func (cmd *DisableServiceAccess) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &cliFlags.StringFlag{ShortName: "p", Usage: T("Disable access to a specified service plan")}
	fs["o"] = &cliFlags.StringFlag{ShortName: "o", Usage: T("Disable access for a specified organization")}

	return command_registry.CommandMetadata{
		Name:        "disable-service-access",
		Description: T("Disable access to a service or service plan for one or all orgs"),
		Usage:       "CF_NAME disable-service-access SERVICE [-p PLAN] [-o ORG]",
		Flags:       fs,
	}
}

func (cmd *DisableServiceAccess) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("disable-service-access"))
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *DisableServiceAccess) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.actor = deps.ServicePlanHandler
	cmd.tokenRefresher = deps.RepoLocator.GetAuthenticationRepository()
	return cmd
}

func (cmd *DisableServiceAccess) Execute(c flags.FlagContext) {
	_, err := cmd.tokenRefresher.RefreshAuthToken()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	serviceName := c.Args()[0]
	planName := c.String("p")
	orgName := c.String("o")

	if planName != "" && orgName != "" {
		cmd.disablePlanAndOrgForService(serviceName, planName, orgName)
	} else if planName != "" {
		cmd.disableSinglePlanForService(serviceName, planName)
	} else if orgName != "" {
		cmd.disablePlansForSingleOrgForService(serviceName, orgName)
	} else {
		cmd.disableServiceForAll(serviceName)
	}

	cmd.ui.Say(T("OK"))
}

func (cmd *DisableServiceAccess) disableServiceForAll(serviceName string) {
	cmd.ui.Say(T("Disabling access to all plans of service {{.ServiceName}} for all orgs as {{.UserName}}...", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName), "UserName": terminal.EntityNameColor(cmd.config.Username())}))
	allPlansAlreadySet, err := cmd.actor.UpdateAllPlansForService(serviceName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansAlreadySet {
		cmd.ui.Say(T("All plans of the service are already inaccessible for all orgs"))
	}
}

func (cmd *DisableServiceAccess) disablePlanAndOrgForService(serviceName string, planName string, orgName string) {
	cmd.ui.Say(T("Disabling access to plan {{.PlanName}} of service {{.ServiceName}} for org {{.OrgName}} as {{.Username}}...", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName), "OrgName": terminal.EntityNameColor(orgName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	planOriginalAccess, err := cmd.actor.UpdatePlanAndOrgForService(serviceName, planName, orgName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.None {
		cmd.ui.Say(T("The plan is already inaccessible for this org"))
	} else if planOriginalAccess != actors.Limited {
		cmd.ui.Say(T("No action taken.  You must disable access to the {{.PlanName}} plan of {{.ServiceName}} service for all orgs and then grant access for all orgs except the {{.OrgName}} org.",
			map[string]interface{}{
				"PlanName":    terminal.EntityNameColor(planName),
				"ServiceName": terminal.EntityNameColor(serviceName),
				"OrgName":     terminal.EntityNameColor(orgName),
			}))
	}
	return
}

func (cmd *DisableServiceAccess) disableSinglePlanForService(serviceName string, planName string) {
	cmd.ui.Say(T("Disabling access of plan {{.PlanName}} for service {{.ServiceName}} as {{.Username}}...", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	planOriginalAccess, err := cmd.actor.UpdateSinglePlanForService(serviceName, planName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.None {
		cmd.ui.Say(T("The plan is already inaccessible for all orgs"))
	}
	return
}

func (cmd *DisableServiceAccess) disablePlansForSingleOrgForService(serviceName string, orgName string) {
	cmd.ui.Say(T("Disabling access to all plans of service {{.ServiceName}} for the org {{.OrgName}} as {{.Username}}...", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName), "OrgName": terminal.EntityNameColor(orgName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	serviceAccess, err := cmd.actor.FindServiceAccess(serviceName, orgName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	if serviceAccess == actors.AllPlansArePublic {
		cmd.ui.Say(T("No action taken.  You must disable access to all plans of {{.ServiceName}} service for all orgs and then grant access for all orgs except the {{.OrgName}} org.",
			map[string]interface{}{
				"ServiceName": terminal.EntityNameColor(serviceName),
				"OrgName":     terminal.EntityNameColor(orgName),
			}))
		return
	}

	allPlansWereSet, err := cmd.actor.UpdateOrgForService(serviceName, orgName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansWereSet {
		cmd.ui.Say(T("All plans of the service are already inaccessible for this org"))
	}
}
