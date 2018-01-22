package buildpack

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
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
	fs["s"] = &flags.StringFlag{ShortName: "s", Usage: T("Specify stack to disambiguate buildpacks with the same name")}

	return commandregistry.CommandMetadata{
		Name:        "delete-buildpack",
		Description: T("Delete a buildpack"),
		Usage: []string{
			T("CF_NAME delete-buildpack BUILDPACK [-f] [-s STACK]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteBuildpack) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
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

	return reqs, nil
}

func (cmd *DeleteBuildpack) Execute(c flags.FlagContext) error {
	var (
		err       error
		buildpack models.Buildpack
	)

	buildpackName := c.Args()[0]
	stack := c.String("s")

	force := c.Bool("f")

	if !force {
		answer := cmd.ui.ConfirmDelete("buildpack", buildpackName)
		if !answer {
			return nil
		}
	}

	if stack == "" {
		buildpack, err = cmd.buildpackRepo.FindByName(buildpackName)
		cmd.ui.Say(T("Deleting buildpack {{.BuildpackName}}...", map[string]interface{}{"BuildpackName": terminal.EntityNameColor(buildpackName)}))
	} else {
		buildpack, err = cmd.buildpackRepo.FindByNameAndStack(buildpackName, stack)
		cmd.ui.Say(T("Deleting buildpack {{.BuildpackName}} with stack {{.Stack}}...", map[string]interface{}{"BuildpackName": terminal.EntityNameColor(buildpackName), "Stack": terminal.EntityNameColor(stack)}))
	}

	switch err.(type) {
	case nil: //do nothing
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Buildpack {{.BuildpackName}} does not exist.", map[string]interface{}{"BuildpackName": buildpackName}))
		return nil
	case *errors.AmbiguousModelError:
		return fmt.Errorf("%s Specify the stack (using -s) to disambiguate.", err.Error())
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
