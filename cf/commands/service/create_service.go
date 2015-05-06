package service

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	cli_json "github.com/cloudfoundry/cli/json"
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
			flag_helpers.NewStringFlag("c", T("Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")),
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
	params := c.String("c")

	paramsMap := make(map[string]interface{})
	paramsMap, err := cmd.parseArbitraryParams(params)
	if err != nil && params != "" {
		cmd.ui.Failed(T("Invalid JSON provided in -c argument"))
	}

	cmd.ui.Say(T("Creating service instance {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceInstanceName),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	plan, err := cmd.CreateService(serviceName, planName, serviceInstanceName, paramsMap)

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

func (cmd CreateService) CreateService(serviceName, planName, serviceInstanceName string, params map[string]interface{}) (models.ServicePlanFields, error) {
	offerings, apiErr := cmd.serviceBuilder.GetServicesByNameForSpaceWithPlans(cmd.config.SpaceFields().Guid, serviceName)
	if apiErr != nil {
		return models.ServicePlanFields{}, apiErr
	}

	plan, apiErr := findPlanFromOfferings(offerings, planName)
	if apiErr != nil {
		return plan, apiErr
	}

	apiErr = cmd.serviceRepo.CreateServiceInstance(serviceInstanceName, plan.Guid, params)
	return plan, apiErr
}

func (cmd CreateService) parseArbitraryParams(paramsFileOrJson string) (map[string]interface{}, error) {
	var paramsMap map[string]interface{}
	var err error

	paramsMap, err = cli_json.ParseJsonHash(paramsFileOrJson)
	if err != nil {
		paramsMap = make(map[string]interface{})
		err = json.Unmarshal([]byte(paramsFileOrJson), &paramsMap)
		if err != nil && paramsFileOrJson != "" {
			return nil, err
		}
	}
	return paramsMap, nil
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
