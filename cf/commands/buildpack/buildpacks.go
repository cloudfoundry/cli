package buildpack

import (
	"errors"
	"strconv"

	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListBuildpacks struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func init() {
	commandregistry.Register(&ListBuildpacks{})
}

func (cmd *ListBuildpacks) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "buildpacks",
		Description: T("List all buildpacks"),
		Usage: []string{
			T("CF_NAME buildpacks"),
		},
	}
}

func (cmd *ListBuildpacks) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *ListBuildpacks) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.buildpackRepo = deps.RepoLocator.GetBuildpackRepository()
	return cmd
}

func (cmd *ListBuildpacks) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Getting buildpacks...\n"))

	table := cmd.ui.Table([]string{"buildpack", T("position"), T("enabled"), T("locked"), T("filename")})
	noBuildpacks := true

	apiErr := cmd.buildpackRepo.ListBuildpacks(func(buildpack models.Buildpack) bool {
		position := ""
		if buildpack.Position != nil {
			position = strconv.Itoa(*buildpack.Position)
		}
		enabled := ""
		if buildpack.Enabled != nil {
			enabled = strconv.FormatBool(*buildpack.Enabled)
		}
		locked := ""
		if buildpack.Locked != nil {
			locked = strconv.FormatBool(*buildpack.Locked)
		}
		table.Add(
			buildpack.Name,
			position,
			enabled,
			locked,
			buildpack.Filename,
		)
		noBuildpacks = false
		return true
	})
	table.Print()

	if apiErr != nil {
		return errors.New(T("Failed fetching buildpacks.\n{{.Error}}", map[string]interface{}{"Error": apiErr.Error()}))
	}

	if noBuildpacks {
		cmd.ui.Say(T("No buildpacks found"))
	}
	return nil
}
