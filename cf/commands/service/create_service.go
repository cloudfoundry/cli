package service

import (
	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateService struct {
	ui             terminal.UI
	config         core_config.Reader
	serviceRepo    api.ServiceRepository
	serviceBuilder service_builder.ServiceBuilder
}

func NewCreateService(ui terminal.UI, config core_config.Reader, serviceRepo api.ServiceRepository, serviceBuilder service_builder.ServiceBuilder) (cmd CreateService) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = serviceRepo
	cmd.serviceBuilder = serviceBuilder
	return
}

func (cmd CreateService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-service",
		ShortName:   "cs",
		Description: T("Create a service instance"),
		Usage: T(`CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE

EXAMPLE:
   CF_NAME create-service dbaas silver mydb

TIP:
   Use 'CF_NAME create-user-provided-service' to make user-provided services available to cf apps`),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "w", Usage: T("Wait for creation confirmation")},
		},
	}
}

func (cmd CreateService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return
}

func (cmd CreateService) Run(c *cli.Context) {
	serviceName := c.Args()[0]
	planName := c.Args()[1]
	serviceInstanceName := c.Args()[2]

	cmd.ui.Say(T("Creating service instance {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceInstanceName),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	var wait bool = c.Bool("w") || false
	plan, err := cmd.CreateService(serviceName, planName, serviceInstanceName, wait)

	switch err.(type) {
	case nil:
		err := printSuccessMessageForServiceInstance(serviceInstanceName, cmd.serviceRepo, cmd.ui)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}

		if !plan.Free {
			cmd.ui.Say("")
			cmd.ui.Say(T("Attention: The plan `{{.PlanName}}` of service `{{.ServiceName}}` is not free.  The instance `{{.ServiceInstanceName}}` will incur a cost.  Contact your administrator if you think this is in error.",
				map[string]interface{}{
					"PlanName":            terminal.EntityNameColor(plan.Name),
					"ServiceName":         terminal.EntityNameColor(serviceName),
					"ServiceInstanceName": terminal.EntityNameColor(serviceInstanceName),
				}))
			cmd.ui.Say("")
		}
	case *errors.ModelAlreadyExistsError:
		cmd.ui.Ok()
		cmd.ui.Warn(err.Error())
	default:
		cmd.ui.Failed(err.Error())
	}
}

func (cmd CreateService) CreateService(serviceName string, planName string, serviceInstanceName string, wait bool) (models.ServicePlanFields, error) {
	offerings, apiErr := cmd.serviceBuilder.GetServicesByNameForSpaceWithPlans(cmd.config.SpaceFields().Guid, serviceName)
	if apiErr != nil {
		return models.ServicePlanFields{}, apiErr
	}

	plan, apiErr := findPlanFromOfferings(offerings, planName)
	if apiErr != nil {
		return plan, apiErr
	}

	apiErr = cmd.serviceRepo.CreateServiceInstance(serviceInstanceName, plan.Guid, wait)
	return plan, apiErr
}

func findPlanFromOfferings(offerings models.ServiceOfferings, name string) (plan models.ServicePlanFields, err error) {
	for _, offering := range offerings {
		for _, plan := range offering.Plans {
			if name == plan.Name {
				return plan, nil
			}
		}
	}

	err = errors.New(T("Could not find plan with name {{.ServicePlanName}}",
		map[string]interface{}{"ServicePlanName": name},
	))
	return
}
