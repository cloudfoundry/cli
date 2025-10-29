package v7

import (
	"strconv"

    "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
    "code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/ui"
)

type BuildpacksCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME buildpacks [--labels SELECTOR] [--lifecycle buildpack|cnb]\n\nEXAMPLES:\n   CF_NAME buildpacks\n   CF_NAME buildpacks --labels 'environment in (production,staging),tier in (backend)'\n   CF_NAME buildpacks --labels 'env=dev,!chargeback-code,tier in (backend,worker)'\n   CF_NAME buildpacks --lifecycle cnb"`
	relatedCommands interface{} `related_commands:"create-buildpack, delete-buildpack, rename-buildpack, update-buildpack"`
	Labels          string      `long:"labels" description:"Selector to filter buildpacks by labels"`
	Lifecycle       string      `long:"lifecycle" description:"Filter buildpacks by lifecycle ('buildpack' or 'cnb')"`
}

func (cmd BuildpacksCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	if cmd.Lifecycle != "" {
		err = command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionBuildpackLifecycleQuery, "--lifecycle")
		if err != nil {
			return err
		}
	}

	cmd.UI.DisplayTextWithFlavor("Getting buildpacks as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	buildpacks, warnings, err := cmd.Actor.GetBuildpacks(cmd.Labels, cmd.Lifecycle)
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

func (cmd BuildpacksCommand) displayTable(buildpacks []resources.Buildpack) {
	if len(buildpacks) > 0 {
		var keyValueTable = [][]string{
			{"position", "name", "stack", "enabled", "locked", "state", "filename", "lifecycle"},
		}

		for _, buildpack := range buildpacks {
			keyValueTable = append(keyValueTable, []string{
				strconv.Itoa(buildpack.Position.Value),
				buildpack.Name,
				buildpack.Stack,
				strconv.FormatBool(buildpack.Enabled.Value),
				strconv.FormatBool(buildpack.Locked.Value),
				buildpack.State,
				buildpack.Filename,
				buildpack.Lifecycle,
			})
		}

		cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
