package service

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	"github.com/cloudfoundry/cli/json"
)

type CreateService struct {
	ui             terminal.UI
	config         core_config.Reader
	serviceRepo    api.ServiceRepository
	serviceBuilder service_builder.ServiceBuilder
}

func init() {
	command_registry.Register(&CreateService{})
}

func (cmd *CreateService) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["c"] = &cliFlags.StringFlag{ShortName: "c", Usage: T("Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")}
	fs["t"] = &cliFlags.StringFlag{ShortName: "t", Usage: T("User provided tags")}

	baseUsage := T("CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-t TAGS]")
	paramsUsage := T(`   Optionally provide service-specific configuration parameters in a valid JSON object in-line:

   CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c '{"name":"value","name":"value"}'

   Optionally provide a file containing service-specific configuration parameters in a valid JSON object.
   The path to the parameters file can be an absolute or relative path to a file:

   CF_NAME create-service SERVICE PLAN SERVICE_INSTANCE -c PATH_TO_FILE

   Example of valid JSON object:
   {
      "cluster_nodes": {
         "count": 5,
         "memory_mb": 1024
      }
   }`)
	exampleUsage := T(`EXAMPLE:
   Linux/Mac:
      CF_NAME create-service db-service silver mydb -c '{"ram_gb":4}'

   Windows Command Line:
      CF_NAME create-service db-service silver mydb -c "{\"ram_gb\":4}"

   Windows PowerShell:
      CF_NAME create-service db-service silver mydb -c '{\"ram_gb\":4}'

   CF_NAME create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json

   CF_NAME create-service db-service silver mydb -t "list, of, tags"`)
	tipsUsage := T(`TIP:
   Use 'CF_NAME create-user-provided-service' to make user-provided services available to cf apps`)
	return command_registry.CommandMetadata{
		Name:        "create-service",
		ShortName:   "cs",
		Description: T("Create a service instance"),
		Usage:       strings.Join([]string{baseUsage, paramsUsage, exampleUsage, tipsUsage}, "\n\n"),
		Flags:       fs,
	}
}

func (cmd *CreateService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires service, service plan, service instance as arguments\n\n") + command_registry.Commands.CommandUsage("create-service"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return
}

func (cmd *CreateService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceBuilder = deps.ServiceBuilder
	return cmd
}

func (cmd *CreateService) Execute(c flags.FlagContext) {
	serviceName := c.Args()[0]
	planName := c.Args()[1]
	serviceInstanceName := c.Args()[2]
	params := c.String("c")
	tags := c.String("t")

	tagsList := ui_helpers.ParseTags(tags)

	paramsMap, err := json.ParseJsonFromFileOrString(params)
	if err != nil {
		cmd.ui.Failed(T("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
	}

	cmd.ui.Say(T("Creating service instance {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceInstanceName),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	plan, err := cmd.CreateService(serviceName, planName, serviceInstanceName, paramsMap, tagsList)

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

func (cmd CreateService) CreateService(serviceName, planName, serviceInstanceName string, params map[string]interface{}, tags []string) (models.ServicePlanFields, error) {
	offerings, apiErr := cmd.serviceBuilder.GetServicesByNameForSpaceWithPlans(cmd.config.SpaceFields().Guid, serviceName)
	if apiErr != nil {
		return models.ServicePlanFields{}, apiErr
	}

	plan, apiErr := findPlanFromOfferings(offerings, planName)
	if apiErr != nil {
		return plan, apiErr
	}

	apiErr = cmd.serviceRepo.CreateServiceInstance(serviceInstanceName, plan.Guid, params, tags)
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
