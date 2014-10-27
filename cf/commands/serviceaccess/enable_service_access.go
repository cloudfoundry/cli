package serviceaccess

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type EnableServiceAccess struct {
	ui             terminal.UI
	config         core_config.Reader
	actor          actors.ServicePlanActor
	tokenRefresher authentication.TokenRefresher
}

func NewEnableServiceAccess(ui terminal.UI, config core_config.Reader, actor actors.ServicePlanActor, tokenRefresher authentication.TokenRefresher) (cmd *EnableServiceAccess) {
	return &EnableServiceAccess{
		ui:             ui,
		config:         config,
		actor:          actor,
		tokenRefresher: tokenRefresher,
	}
}

func (cmd *EnableServiceAccess) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *EnableServiceAccess) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "enable-service-access",
		Description: T("Enable access to a service or service plan for one or all orgs"),
		Usage:       "CF_NAME enable-service-access SERVICE [-p PLAN] [-o ORG]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", T("Enable access to a specified service plan")),
			flag_helpers.NewStringFlag("o", T("Enable access for a specified organization")),
		},
	}
}

func (cmd *EnableServiceAccess) Run(c *cli.Context) {
	_, err := cmd.tokenRefresher.RefreshAuthToken()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	serviceName := c.Args()[0]
	planName := c.String("p")
	orgName := c.String("o")

	if planName != "" && orgName != "" {
		cmd.enablePlanAndOrgForService(serviceName, planName, orgName)
	} else if planName != "" {
		cmd.enablePlanForService(serviceName, planName)
	} else if orgName != "" {
		cmd.enableAllPlansForSingleOrgForService(serviceName, orgName)
	} else {
		cmd.enableAllPlansForService(serviceName)
	}
	cmd.ui.Say("OK")
}

func (cmd *EnableServiceAccess) enablePlanAndOrgForService(serviceName string, planName string, orgName string) {
	cmd.ui.Say(T("Enabling access to plan {{.PlanName}} of service {{.ServiceName}} for org {{.OrgName}} as {{.Username}}...", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName), "OrgName": terminal.EntityNameColor(orgName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	planOriginalAccess, err := cmd.actor.UpdatePlanAndOrgForService(serviceName, planName, orgName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.All {
		cmd.ui.Say(T("The plan is already accessible for this org"))
	}
}

func (cmd *EnableServiceAccess) enablePlanForService(serviceName string, planName string) {
	cmd.ui.Say(T("Enabling access of plan {{.PlanName}} for service {{.ServiceName}} as {{.Username}}...", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	planOriginalAccess, err := cmd.actor.UpdateSinglePlanForService(serviceName, planName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.All {
		cmd.ui.Say(T("The plan is already accessible for all orgs"))
	}
}

func (cmd *EnableServiceAccess) enableAllPlansForService(serviceName string) {
	cmd.ui.Say(T("Enabling access to all plans of service {{.ServiceName}} for all orgs as {{.Username}}...", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	allPlansInServicePublic, err := cmd.actor.UpdateAllPlansForService(serviceName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansInServicePublic {
		cmd.ui.Say(T("All plans of the service are already accessible for all orgs"))
	}
}

func (cmd *EnableServiceAccess) enableAllPlansForSingleOrgForService(serviceName string, orgName string) {
	cmd.ui.Say(T("Enabling access to all plans of service {{.ServiceName}} for the org {{.OrgName}} as {{.Username}}...", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName), "OrgName": terminal.EntityNameColor(orgName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	allPlansWereSet, err := cmd.actor.UpdateOrgForService(serviceName, orgName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansWereSet {
		cmd.ui.Say(T("All plans of the service are already accessible for this org"))
	}
}
