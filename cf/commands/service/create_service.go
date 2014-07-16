package service

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateService struct {
	ui          terminal.UI
	config      configuration.Reader
	serviceRepo api.ServiceRepository
}

func NewCreateService(ui terminal.UI, config configuration.Reader, serviceRepo api.ServiceRepository) (cmd CreateService) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd CreateService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-service",
		ShortName:   "cs",
		Description: T("Create a service instance"),
		Usage: T(`CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE

EXAMPLE:
   CF_NAME create-service cleardb spark clear-db-mine

TIP:
   Use 'CF_NAME create-user-provided-service' to make user-provided services available to cf apps`),
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

	cmd.ui.Say(T("Creating service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceInstanceName),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.CreateService(serviceName, planName, serviceInstanceName)

	switch err.(type) {
	case nil:
		cmd.ui.Ok()
	case *errors.ModelAlreadyExistsError:
		cmd.ui.Ok()
		cmd.ui.Warn(err.Error())
	default:
		cmd.ui.Failed(err.Error())
	}
}

func (cmd CreateService) CreateService(serviceName string, planName string, serviceInstanceName string) (apiErr error) {
	offerings, apiErr := cmd.serviceRepo.FindServiceOfferingsForSpaceByLabel(cmd.config.SpaceFields().Guid, serviceName)
	if apiErr != nil {
		return
	}
	plan, apiErr := findPlanFromOfferings(offerings, planName)
	if apiErr != nil {
		return
	}

	return cmd.serviceRepo.CreateServiceInstance(serviceInstanceName, plan.Guid)
}

func findOfferings(offerings []models.ServiceOffering, name string) (matchingOfferings models.ServiceOfferings, err error) {
	for _, offering := range offerings {
		if name == offering.Label {
			matchingOfferings = append(matchingOfferings, offering)
		}
	}

	if len(matchingOfferings) == 0 {
		err = errors.New(T("Could not find any offerings with name {{.ServiceOfferingName}}",
			map[string]interface{}{"ServiceOfferingName": name},
		))
	}
	return
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
