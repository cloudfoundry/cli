package buildpack

import (
	"errors"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type UpdateBuildpack struct {
	ui                terminal.UI
	buildpackRepo     api.BuildpackRepository
	buildpackBitsRepo api.BuildpackBitsRepository
	buildpackReq      requirements.BuildpackRequirement
}

func init() {
	commandregistry.Register(&UpdateBuildpack{})
}

func (cmd *UpdateBuildpack) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &flags.IntFlag{ShortName: "i", Usage: T("The order in which the buildpacks are checked during buildpack auto-detection")}
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Path to directory or zip file")}
	fs["enable"] = &flags.BoolFlag{Name: "enable", Usage: T("Enable the buildpack to be used for staging")}
	fs["disable"] = &flags.BoolFlag{Name: "disable", Usage: T("Disable the buildpack from being used for staging")}
	fs["lock"] = &flags.BoolFlag{Name: "lock", Usage: T("Lock the buildpack to prevent updates")}
	fs["unlock"] = &flags.BoolFlag{Name: "unlock", Usage: T("Unlock the buildpack to enable updates")}

	return commandregistry.CommandMetadata{
		Name:        "update-buildpack",
		Description: T("Update a buildpack"),
		Usage: []string{
			T("CF_NAME update-buildpack BUILDPACK [-p PATH] [-i POSITION] [--enable|--disable] [--lock|--unlock]"),
			T("\n\nTIP:\n"),
			T("   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."),
		},
		Flags: fs,
	}
}

func (cmd *UpdateBuildpack) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("update-buildpack"))
	}

	loginReq := requirementsFactory.NewLoginRequirement()
	cmd.buildpackReq = requirementsFactory.NewBuildpackRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		loginReq,
		cmd.buildpackReq,
	}

	return reqs
}

func (cmd *UpdateBuildpack) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.buildpackRepo = deps.RepoLocator.GetBuildpackRepository()
	cmd.buildpackBitsRepo = deps.RepoLocator.GetBuildpackBitsRepository()
	return cmd
}

func (cmd *UpdateBuildpack) Execute(c flags.FlagContext) error {
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
		return errors.New(T("Cannot specify both {{.Enabled}} and {{.Disabled}}.", map[string]interface{}{
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
		return errors.New(T("Cannot specify both lock and unlock options."))
	}

	path := c.String("p")
	var dir string
	var err error
	if path != "" {
		dir, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	if dir != "" && (lock || unlock) {
		return errors.New(T("Cannot specify buildpack bits and lock/unlock."))
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
		newBuildpack, err := cmd.buildpackRepo.Update(buildpack)
		if err != nil {
			return errors.New(T("Error updating buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
				"Name":  terminal.EntityNameColor(buildpack.Name),
				"Error": err.Error(),
			}))
		}
		buildpack = newBuildpack
	}

	if dir != "" {
		err := cmd.buildpackBitsRepo.UploadBuildpack(buildpack, dir)
		if err != nil {
			return errors.New(T("Error uploading buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
				"Name":  terminal.EntityNameColor(buildpack.Name),
				"Error": err.Error(),
			}))
		}
	}
	cmd.ui.Ok()
	return nil
}
