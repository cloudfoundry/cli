package serviceaccess

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DisableServiceAccess struct {
	ui     terminal.UI
	config configuration.Reader
	actor  actors.ServicePlanActor
}

func NewDisableServiceAccess(ui terminal.UI, config configuration.Reader, actor actors.ServicePlanActor) (cmd *DisableServiceAccess) {
	return &DisableServiceAccess{
		ui:     ui,
		config: config,
		actor:  actor,
	}
}

func (cmd *DisableServiceAccess) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *DisableServiceAccess) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "disable-service-access",
		Description: T("Disable access to a service or service plan for one or all orgs."),
		Usage:       "CF_NAME disable-service-access SERVICE [-p PLAN] [-o ORG]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", T("Disable access to a particular service plan.")),
			flag_helpers.NewStringFlag("o", T("Disable access to a particular organization.")),
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation.")},
		},
	}
}

func (cmd *DisableServiceAccess) Run(c *cli.Context) {
	serviceName := c.Args()[0]
	planName := c.String("p")
	orgName := c.String("o")
	force := c.Bool("f")

	if planName != "" && orgName != "" {
		cmd.DisablePlanAndOrgForService(serviceName, planName, orgName, force)
	} else if planName != "" {
		cmd.DisableSinglePlanForService(serviceName, planName, force)
	}

	cmd.ui.Say("OK")
}

func (cmd *DisableServiceAccess) DisablePlanAndOrgForService(serviceName string, planName string, orgName string, force bool) {
	prompt := fmt.Sprintf(
		"Really disable access to plan %s of service %s for org %s?",
		terminal.EntityNameColor(planName),
		terminal.EntityNameColor(serviceName),
		terminal.EntityNameColor(orgName),
	)

	if !force {
		if !cmd.confirmDisable(prompt) {
			return
		}
	}

	planOriginalAccess, err := cmd.actor.UpdatePlanAndOrgForService(serviceName, planName, orgName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.None {
		cmd.ui.Say("Plan %s of service %s is already inaccessible for all orgs", planName, serviceName)
	} else if planOriginalAccess == actors.Limited {
		cmd.ui.Say("Disabling access to plan %s of service %s for org %s as %s...", planName, serviceName, orgName, cmd.config.Username())
	} else {
		cmd.ui.Say("The plan %s is accessible to all orgs.", planName)
	}
	return
}

func (cmd *DisableServiceAccess) DisableSinglePlanForService(serviceName string, planName string, force bool) {
	planOriginalAccess, err := cmd.actor.UpdateSinglePlanForService(serviceName, planName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.None {
		cmd.ui.Say("Plan %s for service %s is already private", planName, serviceName)
	} else {
		cmd.ui.Say("Disabling access of plan %s for service %s as %s...", planName, serviceName, cmd.config.Username())
	}
	return
}

func (cmd *DisableServiceAccess) confirmDisable(message string) bool {
	result := cmd.ui.Confirm(message)

	if !result {
		cmd.ui.Warn(T("Disable cancelled"))
	}
	return result
}
