package v7

import (
	"encoding/json"
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	log "github.com/sirupsen/logrus"
)

//go:generate counterfeiter . EnvActor

type EnvActor interface {
	GetEnvironmentVariablesByApplicationNameAndSpace(appName string, spaceGUID string) (v7action.EnvironmentVariableGroups, v7action.Warnings, error)
}

type EnvCommand struct {
	RequiredArgs    flag.EnvironmentArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME env APP_NAME"`
	relatedCommands interface{}          `related_commands:"app, apps, set-env, unset-env, running-environment-variable-group, staging-environment-variable-group"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       EnvActor
}

func (cmd *EnvCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd EnvCommand) Execute(_ []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	appName := cmd.RequiredArgs.AppName
	cmd.UI.DisplayTextWithFlavor("Getting env variables for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   appName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayOK()

	envGroups, warnings, err := cmd.Actor.GetEnvironmentVariablesByApplicationNameAndSpace(
		appName,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(envGroups.System) > 0 || len(envGroups.Application) > 0 {
		cmd.UI.DisplayHeader("System-Provided:")
		err = cmd.displaySystem(envGroups.System)
		if err != nil {
			log.Errorln("error formatting system provided:", err)
		}
		if len(envGroups.Application) > 0 {
			cmd.UI.DisplayNewline()
			err = cmd.displaySystem(envGroups.Application)
			if err != nil {
				log.Errorln("error formatting application:", err)
			}
		}
	} else {
		cmd.UI.DisplayText("No system-provided env variables have been set")
	}
	cmd.UI.DisplayNewline()

	if len(envGroups.EnvironmentVariables) > 0 {
		cmd.UI.DisplayHeader("User-Provided:")
		cmd.displayEnvGroup(envGroups.EnvironmentVariables)
	} else {
		cmd.UI.DisplayText("No user-provided env variables have been set")
	}
	cmd.UI.DisplayNewline()

	if len(envGroups.Running) > 0 {
		cmd.UI.DisplayHeader("Running Environment Variable Groups:")
		cmd.displayEnvGroup(envGroups.Running)
	} else {
		cmd.UI.DisplayText("No running env variables have been set")
	}
	cmd.UI.DisplayNewline()

	if len(envGroups.Staging) > 0 {
		cmd.UI.DisplayHeader("Staging Environment Variable Groups:")
		cmd.displayEnvGroup(envGroups.Staging)
	} else {
		cmd.UI.DisplayText("No staging env variables have been set")
	}
	cmd.UI.DisplayNewline()

	return nil
}

func (cmd EnvCommand) displayEnvGroup(group map[string]interface{}) {
	keys := sortKeys(group)

	for _, key := range keys {
		cmd.UI.DisplayText(fmt.Sprintf("%s: %v", key, group[key]))
	}
}

func sortKeys(group map[string]interface{}) []string {
	keys := make([]string, 0, len(group))
	for key := range group {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (cmd EnvCommand) displaySystem(group map[string]interface{}) error {
	for key, val := range group {
		jsonVal, err := json.MarshalIndent(val, "", " ")
		if err != nil {
			return err
		}
		cmd.UI.DisplayText(fmt.Sprintf("%s: %s", key, jsonVal))
	}
	return nil
}
