package buildpack

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteBuildpack struct {
	ui            terminal.UI
	buildpackRepo api.BuildpackRepository
}

func init() {
	commandregistry.Register(&DeleteBuildpack{})
}

func (cmd *DeleteBuildpack) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.buildpackRepo = deps.RepoLocator.GetBuildpackRepository()
	return cmd
}

func (cmd *DeleteBuildpack) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-buildpack",
		Description: T("Delete a buildpack"),
		Usage: []string{
			T("CF_NAME delete-buildpack BUILDPACK [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteBuildpack) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd), "",
		func() bool {
			return len(fc.Args()) != 1
		},
	)

	loginReq := requirementsFactory.NewLoginRequirement()

	reqs := []requirements.Requirement{
		usageReq,
		loginReq,
	}

	return reqs
}

func (cmd *DeleteBuildpack) Execute(c flags.FlagContext) error {
	buildpackName := c.Args()[0]

	force := c.Bool("f")

	if !force {
		answer := cmd.ui.ConfirmDelete("buildpack", buildpackName)
		if !answer {
			return nil
		}
	}

	cmd.ui.Say(T("Deleting buildpack {{.BuildpackName}}...", map[string]interface{}{"BuildpackName": terminal.EntityNameColor(buildpackName)}))
	buildpack, err := cmd.buildpackRepo.FindByName(buildpackName)

	switch err.(type) {
	case nil: //do nothing
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Buildpack {{.BuildpackName}} does not exist.", map[string]interface{}{"BuildpackName": buildpackName}))
		return nil
	default:
		return err

	}

	err = cmd.buildpackRepo.Delete(buildpack.GUID)
	if err != nil {
		return errors.New(T("Error deleting buildpack {{.Name}}\n{{.Error}}", map[string]interface{}{
			"Name":  terminal.EntityNameColor(buildpack.Name),
			"Error": err.Error(),
		}))
	}

	cmd.ui.Ok()
	return nil
}
