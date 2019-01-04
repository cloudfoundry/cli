package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . BuildpacksActor

type BuildpacksActor interface {
	GetBuildpacks() ([]v7action.Buildpack, v7action.Warnings, error)
}

type BuildpacksCommand struct {
	usage           interface{} `usage:"CF_NAME buildpacks"`
	relatedCommands interface{} `related_commands:"push"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       BuildpacksActor
}

func (cmd *BuildpacksCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)

	return nil
}

func (cmd BuildpacksCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting buildpacks as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	buildpacks, warnings, err := cmd.Actor.GetBuildpacks()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if len(buildpacks) == 0 {
		cmd.UI.DisplayTextWithFlavor("No buildpacks found")
	} else {
		cmd.displayTable(buildpacks)
	}
	return nil
}

func (cmd BuildpacksCommand) displayTable(buildpacks []v7action.Buildpack) {
	if len(buildpacks) > 0 {
		var keyValueTable = [][]string{
			{"position", "name", "stack", "enabled", "locked", "filename"},
		}
		for _, buildpack := range buildpacks {
			keyValueTable = append(keyValueTable, []string{strconv.Itoa(buildpack.Position), buildpack.Name, buildpack.Stack, strconv.FormatBool(buildpack.Enabled), strconv.FormatBool(buildpack.Locked), buildpack.Filename})
		}

		cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
