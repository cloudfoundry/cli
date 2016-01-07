package service

import (
	"encoding/json"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type UpdateUserProvidedService struct {
	ui                              terminal.UI
	config                          core_config.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
	serviceInstanceReq              requirements.ServiceInstanceRequirement
}

func init() {
	command_registry.Register(&UpdateUserProvidedService{})
}

func (cmd *UpdateUserProvidedService) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &cliFlags.StringFlag{ShortName: "p", Usage: T("Credentials")}
	fs["l"] = &cliFlags.StringFlag{ShortName: "l", Usage: T("Syslog Drain Url")}

	return command_registry.CommandMetadata{
		Name:        "update-user-provided-service",
		ShortName:   "uups",
		Description: T("Update user-provided service instance name value pairs"),
		Usage: T(`CF_NAME update-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG-DRAIN-URL]'

EXAMPLE:
   CF_NAME update-user-provided-service my-db-mine -p '{"username":"admin","password":"pa55woRD"}'
   CF_NAME update-user-provided-service my-drain-service -l syslog://example.com`),
		Flags: fs,
	}
}

func (cmd *UpdateUserProvidedService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("update-user-provided-service"))
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.serviceInstanceReq,
	}
	return reqs, nil
}

func (cmd *UpdateUserProvidedService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userProvidedServiceInstanceRepo = deps.RepoLocator.GetUserProvidedServiceInstanceRepository()
	return cmd
}

func (cmd *UpdateUserProvidedService) Execute(c flags.FlagContext) {
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
	cmd.ui.Say(T("TIP: Use '{{.CFRestageCommand}}' for any bound apps to ensure your env variable changes take effect",
		map[string]interface{}{
			"CFRestageCommand": terminal.CommandColor(cf.Name() + " restage"),
		}))

	if params == "" && drainUrl == "" {
		cmd.ui.Warn(T("No flags specified. No changes were made."))
	}
}
