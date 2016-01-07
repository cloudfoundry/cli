package space

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type DeleteSpace struct {
	ui        terminal.UI
	config    core_config.ReadWriter
	spaceRepo spaces.SpaceRepository
	spaceReq  requirements.SpaceRequirement
}

func init() {
	command_registry.Register(&DeleteSpace{})
}

func (cmd *DeleteSpace) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &cliFlags.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")}

	return command_registry.CommandMetadata{
		Name:        "delete-space",
		Description: T("Delete a space"),
		Usage:       T("CF_NAME delete-space SPACE [-f]"),
		Flags:       fs,
	}
}

func (cmd *DeleteSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("delete-space"))
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
		cmd.spaceReq,
	}
	return
}

func (cmd *DeleteSpace) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}
func (cmd *DeleteSpace) Execute(c flags.FlagContext) {
	spaceName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("space"), spaceName) {
			return
		}
	}

	cmd.ui.Say(T("Deleting space {{.TargetSpace}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetSpace": terminal.EntityNameColor(spaceName),
			"TargetOrg":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	space := cmd.spaceReq.GetSpace()

	apiErr := cmd.spaceRepo.Delete(space.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	if cmd.config.SpaceFields().Guid == space.Guid {
		cmd.config.SetSpaceFields(models.SpaceFields{})
		cmd.ui.Say(T("TIP: No space targeted, use '{{.CfTargetCommand}}' to target a space",
			map[string]interface{}{"CfTargetCommand": cf.Name() + " target -s"}))
	}

	return
}
