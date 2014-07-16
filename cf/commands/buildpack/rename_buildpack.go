package buildpack

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RenameBuildpack struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func NewRenameBuildpack(ui terminal.UI, repo api.BuildpackRepository) (cmd *RenameBuildpack) {
	cmd = new(RenameBuildpack)
	cmd.ui = ui
	cmd.buildpackRepo = repo
	return
}

func (cmd *RenameBuildpack) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-buildpack",
		Description: T("Rename a buildpack"),
		Usage:       T("CF_NAME rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME"),
	}
}

func (cmd *RenameBuildpack) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *RenameBuildpack) Run(c *cli.Context) {
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
