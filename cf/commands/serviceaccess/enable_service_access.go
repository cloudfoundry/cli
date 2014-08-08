package serviceaccess

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type EnableServiceAccess struct {
	ui     terminal.UI
	config configuration.Reader
	actor  actors.ServicePlanActor
}

func NewEnableServiceAccess(ui terminal.UI, config configuration.Reader, actor actors.ServicePlanActor) (cmd *EnableServiceAccess) {
	return &EnableServiceAccess{
		ui:     ui,
		config: config,
		actor:  actor,
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
		Description: "Enable access to a service or service plan for one or all orgs",
		Usage:       "CF_NAME enable-service-access SERVICE [-p PLAN] [-o ORG]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", "Enable access to a particular service plan"),
			flag_helpers.NewStringFlag("o", "Enable access to a particular organization"),
		},
	}
}

func (cmd *EnableServiceAccess) Run(c *cli.Context) {
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
	planOriginalAccess, err := cmd.actor.UpdatePlanAndOrgForService(serviceName, planName, orgName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.All {
		cmd.ui.Say("The plan is already accessible for this org")
	} else {
		cmd.ui.Say("Enabling access to plan %s of service %s for org %s as %s...", terminal.EntityNameColor(planName), terminal.EntityNameColor(serviceName), terminal.EntityNameColor(orgName), terminal.EntityNameColor(cmd.config.Username()))
	}
}

func (cmd *EnableServiceAccess) enablePlanForService(serviceName string, planName string) {
	planOriginalAccess, err := cmd.actor.UpdateSinglePlanForService(serviceName, planName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.All {
		cmd.ui.Say("The plan is already accessible for all orgs")
	} else {
		cmd.ui.Say("Enabling access of plan %s for service %s as %s...", terminal.EntityNameColor(planName), terminal.EntityNameColor(serviceName), terminal.EntityNameColor(cmd.config.Username()))
	}
}

func (cmd *EnableServiceAccess) enableAllPlansForService(serviceName string) {
	allPlansInServicePublic, err := cmd.actor.UpdateAllPlansForService(serviceName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansInServicePublic {
		cmd.ui.Say("All plans of the service are already accessible for all orgs")
	} else {
		cmd.ui.Say("Enabling access to all plans of service %s for all orgs as %s...", terminal.EntityNameColor(serviceName), terminal.EntityNameColor(cmd.config.Username()))
	}
}

func (cmd *EnableServiceAccess) enableAllPlansForSingleOrgForService(serviceName string, orgName string) {
	allPlansWereSet, err := cmd.actor.UpdateOrgForService(serviceName, orgName, true)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansWereSet {
		cmd.ui.Say("All plans of the service are already accessible for the org")
	} else {
		cmd.ui.Say("Enabling access to all plans of service %s for the org %s as %s...", terminal.EntityNameColor(serviceName), terminal.EntityNameColor(orgName), terminal.EntityNameColor(cmd.config.Username()))
	}
}
