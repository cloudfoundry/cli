package v7

import (
	"encoding/json"
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/command/flag"
	log "github.com/sirupsen/logrus"
)

type EnvCommand struct {
	command.BaseCommand

	RequiredArgs    flag.EnvironmentArgs `positional-args:"yes"`
	usage           interface{}          `usage:"CF_NAME env APP_NAME"`
	relatedCommands interface{}          `related_commands:"app, apps, set-env, unset-env, running-environment-variable-group, staging-environment-variable-group"`
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
