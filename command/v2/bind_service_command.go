package v2

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . BindServiceActor

type BindServiceActor interface {
	BindServiceBySpace(appName string, ServiceInstanceName string, spaceGUID string, parameters map[string]interface{}) (v2action.Warnings, error)
}

type BindServiceCommand struct {
	RequiredArgs     flag.BindServiceArgs          `positional-args:"yes"`
	ParametersAsJSON flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	usage            interface{}                   `usage:"CF_NAME bind-service APP_NAME SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]\n\n   Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n\n   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n\n   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. \n   The path to the parameters file can be an absolute or relative path to a file.\n   CF_NAME bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE\n\n   Example of valid JSON object:\n   {\n      \"permissions\": \"read-only\"\n   }\n\nEXAMPLES:\n   Linux/Mac:\n      CF_NAME bind-service myapp mydb -c '{\"permissions\":\"read-only\"}'\n\n   Windows Command Line:\n      CF_NAME bind-service myapp mydb -c \"{\\\"permissions\\\":\\\"read-only\\\"}\"\n\n   Windows PowerShell:\n      CF_NAME bind-service myapp mydb -c '{\\\"permissions\\\":\\\"read-only\\\"}'\n\n   CF_NAME bind-service myapp mydb -c ~/workspace/tmp/instance_config.json"`
	relatedCommands  interface{}                   `related_commands:"services"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       BindServiceActor
}

func (cmd *BindServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd BindServiceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Binding service {{.ServiceName}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...", map[string]interface{}{
		"ServiceName": cmd.RequiredArgs.ServiceInstanceName,
		"AppName":     cmd.RequiredArgs.AppName,
		"OrgName":     cmd.Config.TargetedOrganization().Name,
		"SpaceName":   cmd.Config.TargetedSpace().Name,
		"CurrentUser": user.Name,
	})

	warnings, err := cmd.Actor.BindServiceBySpace(cmd.RequiredArgs.AppName, cmd.RequiredArgs.ServiceInstanceName, cmd.Config.TargetedSpace().GUID, cmd.ParametersAsJSON)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, isTakenError := err.(ccerror.ServiceBindingTakenError); isTakenError {
			cmd.UI.DisplayText("App {{.AppName}} is already bound to {{.ServiceName}}.", map[string]interface{}{
				"AppName":     cmd.RequiredArgs.AppName,
				"ServiceName": cmd.RequiredArgs.ServiceInstanceName,
			})
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayText("TIP: Use '{{.CFCommand}} {{.AppName}}' to ensure your env variable changes take effect", map[string]interface{}{
		"CFCommand": fmt.Sprintf("%s restage", cmd.Config.BinaryName()),
		"AppName":   cmd.RequiredArgs.AppName,
	})

	return nil
}
