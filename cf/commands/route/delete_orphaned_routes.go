package route

import (
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type DeleteOrphanedRoutes struct {
	ui        terminal.UI
	spaceRepo spaces.SpaceRepository
	config    coreconfig.Reader
}

func init() {
	commandregistry.Register(&DeleteOrphanedRoutes{})
}

func (cmd *DeleteOrphanedRoutes) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-orphaned-routes",
		Description: T("Delete all orphaned routes (i.e. those that are not mapped to an app)"),
		Usage: []string{
			T("CF_NAME delete-orphaned-routes [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteOrphanedRoutes) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
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

	return reqs, nil
}

func (cmd *DeleteOrphanedRoutes) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *DeleteOrphanedRoutes) Execute(c flags.FlagContext) error {
	force := c.Bool("f")
	if !force {
		response := cmd.ui.Confirm(T("Really delete orphaned routes?{{.Prompt}}",
			map[string]interface{}{"Prompt": terminal.PromptColor(">")}))

		if !response {
			return nil
		}
	}

	err := cmd.spaceRepo.DeleteUnmappedRoutes(cmd.config.SpaceFields().GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
