package service

import (
	"encoding/json"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type UpdateUserProvidedService struct {
	ui                              terminal.UI
	config                          configuration.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
	serviceInstanceReq              requirements.ServiceInstanceRequirement
}

func NewUpdateUserProvidedService(ui terminal.UI, config configuration.Reader, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (cmd *UpdateUserProvidedService) {
	cmd = new(UpdateUserProvidedService)
	cmd.ui = ui
	cmd.config = config
	cmd.userProvidedServiceInstanceRepo = userProvidedServiceInstanceRepo
	return
}

func (cmd *UpdateUserProvidedService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-user-provided-service",
		ShortName:   "uups",
		Description: T("Update user-provided service instance name value pairs"),
		Usage: T(`CF_NAME update-user-provided-service SERVICE_INSTANCE [-p PARAMETERS] [-l SYSLOG-DRAIN-URL]'

EXAMPLE:
   CF_NAME update-user-provided-service oracle-db-mine -p '{"username":"admin","password":"pa55woRD"}'
   CF_NAME update-user-provided-service my-drain-service -l syslog://example.com`),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", T("Parameters")),
			flag_helpers.NewStringFlag("l", T("Syslog Drain Url")),
		},
	}
}

func (cmd *UpdateUserProvidedService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.serviceInstanceReq,
	}

	return
}

func (cmd *UpdateUserProvidedService) Run(c *cli.Context) {

	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	if !serviceInstance.IsUserProvided() {
		cmd.ui.Failed(T("Service Instance is not user provided"))
		return
	}

	drainUrl := c.String("l")
	params := c.String("p")

	paramsMap := make(map[string]interface{})
	if params != "" {

		err := json.Unmarshal([]byte(params), &paramsMap)
		if err != nil {
			cmd.ui.Failed(T("JSON is invalid: {{.ErrorDescription}}", map[string]interface{}{"ErrorDescription": err.Error()}))
			return
		}
	}

	cmd.ui.Say(T("Updating user provided service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceInstance.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstance.Params = paramsMap
	serviceInstance.SysLogDrainUrl = drainUrl

	apiErr := cmd.userProvidedServiceInstanceRepo.Update(serviceInstance.ServiceInstanceFields)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: To make these changes take effect, use '{{.CFUnbindCommand}}' to unbind the service, '{{.CFBindComand}}' to rebind, and then '{{.CFRestageCommand}}' to update the app with the new env variables",
		map[string]interface{}{
			"CFUnbindCommand":  cf.Name() + " unbind-service",
			"CFBindComand":     cf.Name() + " bind-service",
			"CFRestageCommand": cf.Name() + " restage",
		}))

	if params == "" && drainUrl == "" {
		cmd.ui.Warn(T("No flags specified. No changes were made."))
	}
}
