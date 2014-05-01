package buildpack

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/errors"
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

func (command *RenameBuildpack) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-buildpack",
		Description: "Rename a buildpack",
		Usage:       "CF_NAME rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME",
	}
}

func (cmd *RenameBuildpack) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-buildpack")
		return
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *RenameBuildpack) Run(c *cli.Context) {
	buildpackName := c.Args()[0]
	newBuildpackName := c.Args()[1]

	cmd.ui.Say("Renaming buildpack %s to %s...", terminal.EntityNameColor(buildpackName), terminal.EntityNameColor(newBuildpackName))

	buildpack, apiErr := cmd.buildpackRepo.FindByName(buildpackName)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	buildpack.Name = newBuildpackName
	buildpack, apiErr = cmd.buildpackRepo.Update(buildpack)
	if apiErr != nil {
		cmd.ui.Failed("Error renaming buildpack %s\n%s", terminal.EntityNameColor(buildpackName), apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
