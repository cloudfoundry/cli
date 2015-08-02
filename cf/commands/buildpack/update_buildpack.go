package buildpack

import (
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type UpdateBuildpack struct {
	ui                terminal.UI
	buildpackRepo     api.BuildpackRepository
	buildpackBitsRepo api.BuildpackBitsRepository
	buildpackReq      requirements.BuildpackRequirement
}

func init() {
	command_registry.Register(&UpdateBuildpack{})
}

func (cmd *UpdateBuildpack) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &cliFlags.IntFlag{Name: "i", Usage: T("The order in which the buildpacks are checked during buildpack auto-detection")}
	fs["p"] = &cliFlags.StringFlag{Name: "p", Usage: T("Path to directory or zip file")}
	fs["enable"] = &cliFlags.BoolFlag{Name: "enable", Usage: T("Enable the buildpack to be used for staging")}
	fs["disable"] = &cliFlags.BoolFlag{Name: "disable", Usage: T("Disable the buildpack from being used for staging")}
	fs["lock"] = &cliFlags.BoolFlag{Name: "lock", Usage: T("Lock the buildpack to prevent updates")}
	fs["unlock"] = &cliFlags.BoolFlag{Name: "disable", Usage: T("Unlock the buildpack to enable updates")}

	return command_registry.CommandMetadata{
		Name:        "update-buildpack",
		Description: T("Update a buildpack"),
		Usage: T("CF_NAME update-buildpack BUILDPACK [-p PATH] [-i POSITION] [--enable|--disable] [--lock|--unlock]") +
			T("\n\nTIP:\n") + T("   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."),
		Flags: fs,
	}
}

func (cmd *UpdateBuildpack) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("update-buildpack"))
	}

	loginReq := requirementsFactory.NewLoginRequirement()
	cmd.buildpackReq = requirementsFactory.NewBuildpackRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		loginReq,
		cmd.buildpackReq,
	}
	return
}

func (cmd *UpdateBuildpack) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.buildpackRepo = deps.RepoLocator.GetBuildpackRepository()
	cmd.buildpackBitsRepo = deps.RepoLocator.GetBuildpackBitsRepository()
	return cmd
}

func (cmd *UpdateBuildpack) Execute(c flags.FlagContext) {
	buildpack := cmd.buildpackReq.GetBuildpack()

	cmd.ui.Say(T("Updating buildpack {{.BuildpackName}}...", map[string]interface{}{"BuildpackName": terminal.EntityNameColor(buildpack.Name)}))

	updateBuildpack := false

	if c.IsSet("i") {
		position := c.Int("i")

		buildpack.Position = &position
		updateBuildpack = true
	}

	enabled := c.Bool("enable")
	disabled := c.Bool("disable")
	if enabled && disabled {
		cmd.ui.Failed(T("Cannot specify both {{.Enabled}} and {{.Disabled}}.", map[string]interface{}{
			"Enabled":  "enabled",
			"Disabled": "disabled",
		}))
	}

	if enabled {
		buildpack.Enabled = &enabled
		updateBuildpack = true
	}
	if disabled {
		disabled = false
		buildpack.Enabled = &disabled
		updateBuildpack = true
	}

	lock := c.Bool("lock")
	unlock := c.Bool("unlock")
	if lock && unlock {
		cmd.ui.Failed(T("Cannot specify both lock and unlock options."))
		return
	}

	path := c.String("p")
	var dir string
	var err error
	if path != "" {
		dir, err = filepath.Abs(path)
		if err != nil {
			cmd.ui.Failed(err.Error())
			return
		}
	}

	if dir != "" && (lock || unlock) {
		cmd.ui.Failed(T("Cannot specify buildpack bits and lock/unlock."))
	}

	if lock {
		buildpack.Locked = &lock
		updateBuildpack = true
	}
	if unlock {
		unlock = false
		buildpack.Locked = &unlock
		updateBuildpack = true
	}

	if updateBuildpack {
		newBuildpack, apiErr := cmd.buildpackRepo.Update(buildpack)
		if apiErr != nil {
			cmd.ui.Failed(T("Error updating buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
				"Name":  terminal.EntityNameColor(buildpack.Name),
				"Error": apiErr.Error(),
			}))
		}
		buildpack = newBuildpack
	}

	if dir != "" {
		apiErr := cmd.buildpackBitsRepo.UploadBuildpack(buildpack, dir)
		if apiErr != nil {
			cmd.ui.Failed(T("Error uploading buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
				"Name":  terminal.EntityNameColor(buildpack.Name),
				"Error": apiErr.Error(),
			}))
		}
	}
	cmd.ui.Ok()
}
