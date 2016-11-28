package service

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/actors/servicebuilder"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/uihelpers"
	"code.cloudfoundry.org/cli/util/json"
)

type CreateService struct {
	ui             terminal.UI
	config         coreconfig.Reader
	serviceRepo    api.ServiceRepository
	serviceBuilder servicebuilder.ServiceBuilder
}

func init() {
	commandregistry.Register(&CreateService{})
}

func (cmd *CreateService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["c"] = &flags.StringFlag{ShortName: "c", Usage: T("Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")}
	fs["t"] = &flags.StringFlag{ShortName: "t", Usage: T("User provided tags")}

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
	tipsUsage := T(`TIP:
   Use 'CF_NAME create-user-provided-service' to make user-provided services available to CF apps`)
	return commandregistry.CommandMetadata{
		Name:        "create-service",
		ShortName:   "cs",
		Description: T("Create a service instance"),
		Usage: []string{
			baseUsage,
			"\n\n",
			paramsUsage,
			"\n\n",
			tipsUsage,
		},
		Examples: []string{
			fmt.Sprintf("%s:", T(`Linux/Mac`)),
			`   CF_NAME create-service db-service silver mydb -c '{"ram_gb":4}'`,
			``,
			fmt.Sprintf("%s:", T(`Windows Command Line`)),
			`   CF_NAME create-service db-service silver mydb -c "{\"ram_gb\":4}"`,
			``,
			fmt.Sprintf("%s:", T(`Windows PowerShell`)),
			`   CF_NAME create-service db-service silver mydb -c '{\"ram_gb\":4}'`,
			``,
			`CF_NAME create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json`,
			``,
			`CF_NAME create-service db-service silver mydb -t "list, of, tags"`,
		},
		Flags: fs,
	}
}

func (cmd *CreateService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 3 {
		cmd.ui.Failed(T("Incorrect Usage. Requires service, service plan, service instance as arguments\n\n") + commandregistry.Commands.CommandUsage("create-service"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 3)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return reqs, nil
}

func (cmd *CreateService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.serviceBuilder = deps.ServiceBuilder
	return cmd
}

func (cmd *CreateService) Execute(c flags.FlagContext) error {
	serviceName := c.Args()[0]
	planName := c.Args()[1]
	serviceInstanceName := c.Args()[2]
	params := c.String("c")
	tags := c.String("t")

	tagsList := uihelpers.ParseTags(tags)

	paramsMap, err := json.ParseJSONFromFileOrString(params)
	if err != nil {
		return errors.New(T("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
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
			return err
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
		return err
	}
	return nil
}

func (cmd CreateService) CreateService(serviceName, planName, serviceInstanceName string, params map[string]interface{}, tags []string) (models.ServicePlanFields, error) {
	offerings, apiErr := cmd.serviceBuilder.GetServicesByNameForSpaceWithPlans(cmd.config.SpaceFields().GUID, serviceName)
	if apiErr != nil {
		return models.ServicePlanFields{}, apiErr
	}

	plan, apiErr := findPlanFromOfferings(offerings, planName)
	if apiErr != nil {
		return plan, apiErr
	}

	apiErr = cmd.serviceRepo.CreateServiceInstance(serviceInstanceName, plan.GUID, params, tags)
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
