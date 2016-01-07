package commands

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type Target struct {
	ui        terminal.UI
	config    core_config.ReadWriter
	orgRepo   organizations.OrganizationRepository
	spaceRepo spaces.SpaceRepository
}

func init() {
	command_registry.Register(&Target{})
}

func (cmd *Target) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["o"] = &cliFlags.StringFlag{ShortName: "o", Usage: T("organization")}
	fs["s"] = &cliFlags.StringFlag{ShortName: "s", Usage: T("space")}

	return command_registry.CommandMetadata{
		Name:        "target",
		ShortName:   "t",
		Description: T("Set or view the targeted org or space"),
		Usage:       T("CF_NAME target [-o ORG] [-s SPACE]"),
		Flags:       fs,
	}
}

func (cmd *Target) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("target"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewApiEndpointRequirement(),
	}

	if fc.IsSet("o") || fc.IsSet("s") {
		reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	}

	return
}

func (cmd *Target) SetDependency(deps command_registry.Dependency, _ bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *Target) Execute(c flags.FlagContext) {
	orgName := c.String("o")
	spaceName := c.String("s")

	if orgName != "" {
		err := cmd.setOrganization(orgName)
		if err != nil {
			cmd.ui.Failed(err.Error())
		} else if spaceName == "" {
			spaceList, apiErr := cmd.getSpaceList()
			if apiErr == nil && len(spaceList) == 1 {
				cmd.setSpace(spaceList[0].Name)
			}
		}
	}

	if spaceName != "" {
		err := cmd.setSpace(spaceName)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.ShowConfiguration(cmd.config)
	if !cmd.config.IsLoggedIn() {
		cmd.ui.PanicQuietly()
	}
	cmd.ui.NotifyUpdateIfNeeded(cmd.config)
	return
}

func (cmd Target) setOrganization(orgName string) error {
	// setting an org necessarily invalidates any space you had previously targeted
	cmd.config.SetOrganizationFields(models.OrganizationFields{})
	cmd.config.SetSpaceFields(models.SpaceFields{})

	org, apiErr := cmd.orgRepo.FindByName(orgName)
	if apiErr != nil {
		return fmt.Errorf(T("Could not target org.\n{{.ApiErr}}",
			map[string]interface{}{"ApiErr": apiErr.Error()}))
	}

	cmd.config.SetOrganizationFields(org.OrganizationFields)
	return nil
}

func (cmd Target) setSpace(spaceName string) error {
	cmd.config.SetSpaceFields(models.SpaceFields{})

	if !cmd.config.HasOrganization() {
		return errors.New(T("An org must be targeted before targeting a space"))
	}

	space, apiErr := cmd.spaceRepo.FindByName(spaceName)
	if apiErr != nil {
		return fmt.Errorf(T("Unable to access space {{.SpaceName}}.\n{{.ApiErr}}",
			map[string]interface{}{"SpaceName": spaceName, "ApiErr": apiErr.Error()}))
	}

	cmd.config.SetSpaceFields(space.SpaceFields)
	return nil
}

func (cmd Target) getSpaceList() ([]models.Space, error) {
	spaceList := []models.Space{}
	apiErr := cmd.spaceRepo.ListSpaces(
		func(space models.Space) bool {
			spaceList = append(spaceList, space)
			return true
		})
	return spaceList, apiErr
}
