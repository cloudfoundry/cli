package space

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type DeleteSpace struct {
	ui        terminal.UI
	config    coreconfig.ReadWriter
	spaceRepo spaces.SpaceRepository
	spaceReq  requirements.SpaceRequirement
	orgRepo   organizations.OrganizationRepository
}

func init() {
	commandregistry.Register(&DeleteSpace{})
}

func (cmd *DeleteSpace) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}
	fs["o"] = &flags.StringFlag{ShortName: "o", Usage: T("Delete space within specified org")}

	return commandregistry.CommandMetadata{
		Name:        "delete-space",
		Description: T("Delete a space"),
		Usage: []string{
			T("CF_NAME delete-space SPACE [-o ORG] [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-space"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.spaceReq = requirementsFactory.NewSpaceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.String("o") == "" {
		reqs = append(reqs, requirementsFactory.NewTargetedOrgRequirement())
		reqs = append(reqs, cmd.spaceReq)
	}

	return reqs, nil
}

func (cmd *DeleteSpace) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	return cmd
}
func (cmd *DeleteSpace) Execute(c flags.FlagContext) error {
	var space models.Space

	spaceName := c.Args()[0]
	orgName := c.String("o")

	if orgName == "" {
		space = cmd.spaceReq.GetSpace()

		orgName = cmd.config.OrganizationFields().Name
	} else {
		org, err := cmd.orgRepo.FindByName(orgName)
		if err != nil {
			return err
		}

		space, err = cmd.spaceRepo.FindByNameInOrg(spaceName, org.GUID)
		if err != nil {
			return err
		}
	}

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("space"), spaceName) {
			return nil
		}
	}

	cmd.ui.Say(T("Deleting space {{.TargetSpace}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetSpace": terminal.EntityNameColor(spaceName),
			"TargetOrg":   terminal.EntityNameColor(orgName),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.spaceRepo.Delete(space.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()

	if cmd.config.SpaceFields().GUID == space.GUID {
		cmd.config.SetSpaceFields(models.SpaceFields{})
		cmd.ui.Say(T("TIP: No space targeted, use '{{.CfTargetCommand}}' to target a space.",
			map[string]interface{}{"CfTargetCommand": cf.Name + " target -s"}))
	}

	return nil
}
