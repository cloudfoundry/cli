package space

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteSpace struct {
	ui        terminal.UI
	config    coreconfig.ReadWriter
	spaceRepo spaces.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func init() {
	commandregistry.Register(&DeleteSpace{})
}

func (cmd *DeleteSpace) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-space",
		Description: T("Delete a space"),
		Usage: []string{
			T("CF_NAME delete-space SPACE [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-space"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}

	return reqs
}

func (cmd *DeleteSpace) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}
func (cmd *DeleteSpace) Execute(c flags.FlagContext) error {
	spaceName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("space"), spaceName) {
			return nil
		}
	}

	cmd.ui.Say(T("Deleting space {{.TargetSpace}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetSpace": terminal.EntityNameColor(spaceName),
			"TargetOrg":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	space := cmd.spaceReq.GetSpace()

	err := cmd.spaceRepo.Delete(space.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()

	if cmd.config.SpaceFields().GUID == space.GUID {
		cmd.config.SetSpaceFields(models.SpaceFields{})
		cmd.ui.Say(T("TIP: No space targeted, use '{{.CfTargetCommand}}' to target a space",
			map[string]interface{}{"CfTargetCommand": cf.Name + " target -s"}))
	}

	return nil
}
