package service

import (
	"strings"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/json"
	"github.com/codegangsta/cli"
)

type BindService struct {
	ui                 terminal.UI
	config             core_config.Reader
	serviceBindingRepo api.ServiceBindingRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

type ServiceBinder interface {
	BindApplication(app models.Application, serviceInstance models.ServiceInstance, paramsMap map[string]interface{}) (apiErr error)
}

func NewBindService(ui terminal.UI, config core_config.Reader, serviceBindingRepo api.ServiceBindingRepository) (cmd *BindService) {
	cmd = new(BindService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceBindingRepo = serviceBindingRepo
	return
}

func (cmd *BindService) Metadata() command_metadata.CommandMetadata {
	baseUsage := T("CF_NAME bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]")
	paramsUsage := T(`   Optionally provide service-specific configuration parameters in a valid JSON object in-line:

   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c '{"name":"value","name":"value"}'

   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. 
   The path to the parameters file can be an absolute or relative path to a file.
   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE

   Example of valid JSON object:
   {
      "permissions": "read-only"
   }`)
	exampleUsage := T(`EXAMPLE:
   Linux/Mac:
      CF_NAME bind-service myapp mydb -c '{"permissions":"read-only"}'

   Windows Command Line:
      CF_NAME bind-service myapp mydb -c "{\"permissions\":\"read-only\"}"

   Windows PowerShell:
      CF_NAME bind-service myapp mydb -c '{\"permissions\":\"read-only\"}'
	
   CF_NAME bind-service myapp mydb -c ~/workspace/tmp/instance_config.json`)

	return command_metadata.CommandMetadata{
		Name:        "bind-service",
		ShortName:   "bs",
		Description: T("Bind a service instance to an app"),
		Usage:       strings.Join([]string{baseUsage, paramsUsage, exampleUsage}, "\n\n"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("c", T("Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")),
		},
	}
}

func (cmd *BindService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	serviceName := c.Args()[1]

	if cmd.appReq == nil {
		cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	} else {
		cmd.appReq.SetApplicationName(c.Args()[0])
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(), cmd.appReq, cmd.serviceInstanceReq}
	return
}

func (cmd *BindService) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	params := c.String("c")

	paramsMap, err := json.ParseJsonFromFileOrString(params)
	if err != nil {
		cmd.ui.Failed(T("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
	}

	cmd.ui.Say(T("Binding service {{.ServiceInstanceName}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"AppName":             terminal.EntityNameColor(app.Name),
			"OrgName":             terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":           terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser":         terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.BindApplication(app, serviceInstance, paramsMap)
	if err != nil {
		if httperr, ok := err.(errors.HttpError); ok && httperr.ErrorCode() == errors.APP_ALREADY_BOUND {
			cmd.ui.Ok()
			cmd.ui.Warn(T("App {{.AppName}} is already bound to {{.ServiceName}}.",
				map[string]interface{}{
					"AppName":     app.Name,
					"ServiceName": serviceInstance.Name,
				}))
			return
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: Use '{{.CFCommand}} {{.AppName}}' to ensure your env variable changes take effect",
		map[string]interface{}{"CFCommand": terminal.CommandColor(cf.Name() + " restage"), "AppName": app.Name}))
}

func (cmd *BindService) BindApplication(app models.Application, serviceInstance models.ServiceInstance, paramsMap map[string]interface{}) (apiErr error) {
	apiErr = cmd.serviceBindingRepo.Create(serviceInstance.Guid, app.Guid, paramsMap)
	return
}
