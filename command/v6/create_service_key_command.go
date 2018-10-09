package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . CreateServiceKeyActor

type CreateServiceKeyActor interface {
	CreateServiceKey(serviceInstanceName, keyName, spaceGUID string, parameters map[string]interface{}) (v2action.ServiceKey, v2action.Warnings, error)
}

type CreateServiceKeyCommand struct {
	RequiredArgs     flag.ServiceInstanceKey       `positional-args:"yes"`
	ParametersAsJSON flag.JSONOrFileWithValidation `short:"c" description:"Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."`
	usage            interface{}                   `usage:"CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY [-c PARAMETERS_AS_JSON]\n\n   Optionally provide service-specific configuration parameters in a valid JSON object in-line.\n   CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c '{\"name\":\"value\",\"name\":\"value\"}'\n\n   Optionally provide a file containing service-specific configuration parameters in a valid JSON object. The path to the parameters file can be an absolute or relative path to a file.\n   CF_NAME create-service-key SERVICE_INSTANCE SERVICE_KEY -c PATH_TO_FILE\n\n   Example of valid JSON object:\n   {\n      \"permissions\": \"read-only\"\n   }\n\nEXAMPLES:\n   CF_NAME create-service-key mydb mykey -c '{\"permissions\":\"read-only\"}'\n   CF_NAME create-service-key mydb mykey -c ~/workspace/tmp/instance_config.json"`
	relatedCommands  interface{}                   `related_commands:"service-key"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateServiceKeyActor
}

func (cmd *CreateServiceKeyCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd CreateServiceKeyCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	serviceInstanceName := cmd.RequiredArgs.ServiceInstance
	keyName := cmd.RequiredArgs.ServiceKey

	cmd.UI.DisplayTextWithFlavor("Creating service key {{.KeyName}} for service instance {{.ServiceInstanceName}} as {{.User}}...",
		map[string]interface{}{
			"KeyName":             keyName,
			"ServiceInstanceName": serviceInstanceName,
			"User":                user.Name,
		})

	_, warnings, err := cmd.Actor.CreateServiceKey(serviceInstanceName, keyName, cmd.Config.TargetedSpace().GUID, cmd.ParametersAsJSON)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, isTakenError := err.(ccerror.ServiceKeyTakenError); isTakenError {
			cmd.UI.DisplayOK()
			cmd.UI.DisplayWarning("Service key {{.KeyName}} already exists", map[string]interface{}{
				"KeyName": keyName,
			})
			return nil
		}
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
