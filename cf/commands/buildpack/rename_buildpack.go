package buildpack

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
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

func (cmd *RenameBuildpack) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires BUILDPACK_NAME, NEW_BUILDPACK_NAME as arguments\n\n") + commandregistry.Commands.CommandUsage("rename-buildpack"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *RenameBuildpack) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.buildpackRepo = deps.RepoLocator.GetBuildpackRepository()
	return cmd
}

func (cmd *RenameBuildpack) Execute(c flags.FlagContext) error {
	buildpackName := c.Args()[0]
	newBuildpackName := c.Args()[1]

	cmd.ui.Say(T("Renaming buildpack {{.OldBuildpackName}} to {{.NewBuildpackName}}...", map[string]interface{}{"OldBuildpackName": terminal.EntityNameColor(buildpackName), "NewBuildpackName": terminal.EntityNameColor(newBuildpackName)}))

	buildpack, err := cmd.buildpackRepo.FindByName(buildpackName)

	if err != nil {
		return err
	}

	buildpack.Name = newBuildpackName
	buildpack, err = cmd.buildpackRepo.Update(buildpack)
	if err != nil {
		return errors.New(T("Error renaming buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
			"Name":  terminal.EntityNameColor(buildpack.Name),
			"Error": err.Error(),
		}))
	}

	cmd.ui.Ok()
	return nil
}
