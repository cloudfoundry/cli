package buildpack

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteBuildpack struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func NewDeleteBuildpack(ui terminal.UI, repo api.BuildpackRepository) (cmd *DeleteBuildpack) {
	cmd = new(DeleteBuildpack)
	cmd.ui = ui
	cmd.buildpackRepo = repo
	return
}

func (cmd *DeleteBuildpack) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-buildpack",
		Description: T("Delete a buildpack"),
		Usage:       T("CF_NAME delete-buildpack BUILDPACK [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd *DeleteBuildpack) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	loginReq := requirementsFactory.NewLoginRequirement()

	reqs = []requirements.Requirement{
		loginReq,
	}

	return
}

func (cmd *DeleteBuildpack) Run(c *cli.Context) {
	buildpackName := c.Args()[0]

	force := c.Bool("f")

	if !force {
		answer := cmd.ui.ConfirmDelete("buildpack", buildpackName)
		if !answer {
			return
		}
	}

	cmd.ui.Say(T("Deleting buildpack {{.BuildpackName}}...", map[string]interface{}{"BuildpackName": terminal.EntityNameColor(buildpackName)}))
	buildpack, apiErr := cmd.buildpackRepo.FindByName(buildpackName)

	switch apiErr.(type) {
	case nil: //do nothing
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Buildpack {{.BuildpackName}} does not exist.", map[string]interface{}{"BuildpackName": buildpackName}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return

	}

	apiErr = cmd.buildpackRepo.Delete(buildpack.Guid)
	if apiErr != nil {
		cmd.ui.Failed(T("Error deleting buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
			"Name":  terminal.EntityNameColor(buildpack.Name),
			"Error": apiErr.Error(),
		}))
	}

	cmd.ui.Ok()
}
