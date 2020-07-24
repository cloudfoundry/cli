package commands

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type Target struct {
	ui        terminal.UI
	config    coreconfig.ReadWriter
	orgRepo   organizations.OrganizationRepository
	spaceRepo spaces.SpaceRepository
}

func init() {
	commandregistry.Register(&Target{})
}

func (cmd *Target) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["o"] = &flags.StringFlag{ShortName: "o", Usage: T("Organization")}
	fs["s"] = &flags.StringFlag{ShortName: "s", Usage: T("Space")}

	return commandregistry.CommandMetadata{
		Name:        "target",
		ShortName:   "t",
		Description: T("Set or view the targeted org or space"),
		Usage: []string{
			T("CF_NAME target [-o ORG] [-s SPACE]"),
		},
		Flags: fs,
	}
}

func (cmd *Target) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewAPIEndpointRequirement(),
	}

	if fc.IsSet("o") || fc.IsSet("s") {
		reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	}

	return reqs, nil
}

func (cmd *Target) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	return cmd
}

func (cmd *Target) Execute(c flags.FlagContext) error {
	orgName := c.String("o")
	spaceName := c.String("s")

	if orgName != "" {
		err := cmd.setOrganization(orgName)
		if err != nil {
			return err
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
			return err
		}
	}

	err := cmd.ui.ShowConfiguration(cmd.config)
	if err != nil {
		return err
	}
	cmd.ui.NotifyUpdateIfNeeded(cmd.config)
	if !cmd.config.IsLoggedIn() {
		return fmt.Errorf("") // Done on purpose, do not redo this in refactor code
	}
	return nil
}

func (cmd Target) setOrganization(orgName string) error {
	// setting an org necessarily invalidates any space you had previously targeted
	cmd.config.SetOrganizationFields(models.OrganizationFields{})
	cmd.config.SetSpaceFields(models.SpaceFields{})

	org, apiErr := cmd.orgRepo.FindByName(orgName)
	if apiErr != nil {
		return fmt.Errorf(T("Could not target org.\n{{.APIErr}}",
			map[string]interface{}{"APIErr": apiErr.Error()}))
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
		return fmt.Errorf(T("Unable to access space {{.SpaceName}}.\n{{.APIErr}}",
			map[string]interface{}{"SpaceName": spaceName, "APIErr": apiErr.Error()}))
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
