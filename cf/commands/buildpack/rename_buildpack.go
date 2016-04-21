package buildpack

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RenameBuildpack struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func init() {
	commandregistry.Register(&RenameBuildpack{})
}

func (cmd *RenameBuildpack) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "rename-buildpack",
		Description: T("Rename a buildpack"),
		Usage: []string{
			T("CF_NAME rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME"),
		},
	}
}

func (cmd *RenameBuildpack) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires BUILDPACK_NAME, NEW_BUILDPACK_NAME as arguments\n\n") + commandregistry.Commands.CommandUsage("rename-buildpack"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *RenameBuildpack) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.buildpackRepo = deps.RepoLocator.GetBuildpackRepository()
	return cmd
}

func (cmd *RenameBuildpack) Execute(c flags.FlagContext) {
	buildpackName := c.Args()[0]
	newBuildpackName := c.Args()[1]

	cmd.ui.Say(T("Renaming buildpack {{.OldBuildpackName}} to {{.NewBuildpackName}}...", map[string]interface{}{"OldBuildpackName": terminal.EntityNameColor(buildpackName), "NewBuildpackName": terminal.EntityNameColor(newBuildpackName)}))

	buildpack, apiErr := cmd.buildpackRepo.FindByName(buildpackName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	buildpack.Name = newBuildpackName
	buildpack, apiErr = cmd.buildpackRepo.Update(buildpack)
	if apiErr != nil {
		cmd.ui.Failed(T("Error renaming buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
			"Name":  terminal.EntityNameColor(buildpack.Name),
			"Error": apiErr.Error(),
		}))
	}

	cmd.ui.Ok()
}
