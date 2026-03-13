package v7

import (
	"fmt"
	"sort"
	"time"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/v7/shared"
	"code.cloudfoundry.org/cli/v8/resources"
)

type CleanupOutdatedServiceBindingsCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName `positional-args:"yes"`
	Force               bool         `long:"force" short:"f" description:"Force deletion without confirmation"`
	KeepLast            *int         `long:"keep-last" description:"Keep the last N service bindings (default: 1)"`
	ServiceInstanceName string       `long:"service-instance" description:"Only delete service bindings for the specified service instance"`
	Wait                bool         `long:"wait" short:"w" description:"Wait for the operation(s) to complete"`

	relatedCommands interface{} `related_commands:"bind-service, unbind-service"`
}

type bindingKey struct {
	AppGUID             string
	ServiceInstanceGUID string
}

func (cmd CleanupOutdatedServiceBindingsCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(true, true); err != nil {
		return err
	}
	if err := cmd.displayIntro(); err != nil {
		return err
	}

	var (
		bindings []resources.ServiceCredentialBinding
		warnings v7action.Warnings
		err      error
	)

	if cmd.ServiceInstanceName == "" {
		bindings, warnings, err = cmd.Actor.ListAppBindings(
			v7action.ListAppBindingParams{
				SpaceGUID: cmd.Config.TargetedSpace().GUID,
				AppName:   cmd.RequiredArgs.AppName,
			})
	} else {
		bindings, warnings, err = cmd.Actor.ListServiceAppBindings(
			v7action.ListServiceAppBindingParams{
				SpaceGUID:           cmd.Config.TargetedSpace().GUID,
				ServiceInstanceName: cmd.ServiceInstanceName,
				AppName:             cmd.RequiredArgs.AppName,
			})
	}
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	keepLast := 1
	if cmd.KeepLast != nil {
		if *cmd.KeepLast > 0 {
			keepLast = *cmd.KeepLast
		} else {
			cmd.UI.DisplayWarning(fmt.Sprintf("Invalid argument for --keep-last: %d. Using default value of 1.", *cmd.KeepLast))
		}
	}

	bindingsToDelete := GetOutdatedServiceBindings(bindings, keepLast)

	if len(bindingsToDelete) == 0 {
		cmd.UI.DisplayText("No outdated service bindings found.")
		cmd.UI.DisplayOK()
		return nil
	} else if len(bindingsToDelete) == 1 {
		cmd.UI.DisplayText("Found 1 outdated service binding.")
	} else {
		cmd.UI.DisplayText(fmt.Sprintf("Found %d outdated service bindings.", len(bindingsToDelete)))
	}

	if !cmd.Force {
		response, promptErr := cmd.UI.DisplayBoolPrompt(false, "Really delete all outdated service bindings?")

		if promptErr != nil {
			return promptErr
		}

		if !response {
			cmd.UI.DisplayText("Outdated service bindings have not been deleted.")
			return nil
		}
	}

	for _, binding := range bindingsToDelete {
		cmd.UI.DisplayText("Deleting service binding {{.BindingGUID}}...", map[string]interface{}{"BindingGUID": binding.GUID})

		stream, warnings, err := cmd.Actor.DeleteServiceAppBinding(v7action.DeleteServiceAppBindingParams{
			ServiceBindingGUID: binding.GUID,
		})
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		completed, err := shared.WaitForResult(stream, cmd.UI, cmd.Wait)
		switch {
		case err != nil:
			return err
		case completed:
			cmd.UI.DisplayOK()
		default:
			si, warnings, err := cmd.Actor.GetServiceInstanceByGUID(binding.ServiceInstanceGUID)
			cmd.UI.DisplayWarnings(warnings)
			if err != nil {
				return err
			}
			cmd.UI.DisplayOK()
			cmd.UI.DisplayText("Unbinding in progress. Use 'cf service {{.ServiceInstanceName}}' to check operation status.", map[string]interface{}{"ServiceInstanceName": si.Name})
		}
	}

	return nil
}

func (cmd CleanupOutdatedServiceBindingsCommand) Usage() string {
	return `CF_NAME cleanup-outdated-service-bindings APP_NAME [--keep-last N] [--service-instance SERVICE_INSTANCE_NAME] [--force] [--wait]`
}

func (cmd CleanupOutdatedServiceBindingsCommand) Examples() string {
	return `
CF_NAME cleanup-outdated-service-bindings myapp
CF_NAME cleanup-outdated-service-bindings myapp --keep-last 2 --service-instance myinstance --wait
`
}

// GetOutdatedServiceBindings returns a list that is sorted by 1. ServiceInstanceGUID and 2. CreatedAt ascending
func GetOutdatedServiceBindings(bindings []resources.ServiceCredentialBinding, keepLast int) []resources.ServiceCredentialBinding {
	bindingGroups := make(map[bindingKey][]resources.ServiceCredentialBinding)
	for _, binding := range bindings {
		key := bindingKey{binding.AppGUID, binding.ServiceInstanceGUID}
		bindingGroups[key] = append(bindingGroups[key], binding)
	}

	parseTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}
		}
		return t
	}

	var outdatedBindings []resources.ServiceCredentialBinding
	for key := range bindingGroups {
		slice := bindingGroups[key]
		sort.Slice(slice, func(i, j int) bool {
			return parseTime(slice[i].CreatedAt).After(parseTime(slice[j].CreatedAt))
		})
		if len(slice) > keepLast {
			outdatedBindings = append(outdatedBindings, slice[keepLast:]...)
		}
	}

	sort.SliceStable(outdatedBindings, func(i, j int) bool {
		if outdatedBindings[i].ServiceInstanceGUID != outdatedBindings[j].ServiceInstanceGUID {
			return outdatedBindings[i].ServiceInstanceGUID < outdatedBindings[j].ServiceInstanceGUID
		}
		return parseTime(outdatedBindings[i].CreatedAt).Before(parseTime(outdatedBindings[j].CreatedAt))
	})

	return outdatedBindings
}

func (cmd CleanupOutdatedServiceBindingsCommand) displayIntro() error {
	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor(
		"Cleaning up outdated service bindings for app {{.AppName}} in org {{.Org}} / space {{.Space}} as {{.User}}...",
		map[string]interface{}{
			"AppName": cmd.RequiredArgs.AppName,
			"User":    user.Name,
			"Space":   cmd.Config.TargetedSpace().Name,
			"Org":     cmd.Config.TargetedOrganization().Name,
		},
	)

	return nil
}
