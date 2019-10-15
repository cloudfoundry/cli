package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . StagingEnvironmentVariableGroupActor

type StagingEnvironmentVariableGroupActor interface {
	GetEnvironmentVariableGroup(group constant.EnvironmentVariableGroupName) (v7action.EnvironmentVariableGroup, v7action.Warnings, error)
}

type StagingEnvironmentVariableGroupCommand struct {
	usage           interface{} `usage:"CF_NAME staging-environment-variable-group"`
	relatedCommands interface{} `related_commands:"env, running-environment-variable-group"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       StagingEnvironmentVariableGroupActor
}

func (cmd *StagingEnvironmentVariableGroupCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd StagingEnvironmentVariableGroupCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting the staging environment variable group as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	envVars, warnings, err := cmd.Actor.GetEnvironmentVariableGroup(constant.StagingEnvironmentVariableGroup)
	cmd.UI.DisplayWarningsV7(warnings)
	if err != nil {
		return err
	}

	if len(envVars) == 0 {
		cmd.UI.DisplayTextWithFlavor("No staging environment variable group has been set.")
	} else {
		cmd.displayTable(envVars)
	}

	return nil
}

func (cmd StagingEnvironmentVariableGroupCommand) displayTable(envVars v7action.EnvironmentVariableGroup) {
	var keyValueTable = [][]string{
		{"variable name", "assigned value"},
	}

	for envVarName, envVarValue := range envVars {
		keyValueTable = append(keyValueTable, []string{
			envVarName,
			envVarValue.Value,
		})
	}

	cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
}
