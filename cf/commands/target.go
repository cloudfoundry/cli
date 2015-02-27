package commands

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Target struct {
	ui           terminal.UI
	config       core_config.ReadWriter
	endpointRepo api.EndpointRepository
	orgRepo      organizations.OrganizationRepository
	spaceRepo    spaces.SpaceRepository
}

func NewTarget(ui terminal.UI,
	config core_config.ReadWriter,
	endpointRepo api.EndpointRepository,
	orgRepo organizations.OrganizationRepository,
	spaceRepo spaces.SpaceRepository) (cmd Target) {

	cmd.ui = ui
	cmd.config = config
	cmd.endpointRepo = endpointRepo
	cmd.orgRepo = orgRepo
	cmd.spaceRepo = spaceRepo

	return
}

func (cmd Target) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "target",
		ShortName:   "t",
		Description: T("Set or view the targeted org, space or api endpoint"),
		Usage:       T("CF_NAME target [-a API] [-o ORG] [-s SPACE]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("a", T("API endpoint (e.g. https://api.example.com)")),
			flag_helpers.NewStringFlag("o", T("organization")),
			flag_helpers.NewStringFlag("s", T("space")),
			cli.BoolFlag{Name: "skip-ssl-validation", Usage: T("Please don't")},
		},
	}
}

func (cmd Target) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 0 {
		err = errors.New(T("incorrect usage"))
		cmd.ui.FailWithUsage(c)
		return
	}

	if c.String("a") == "" {
		reqs = append(reqs, requirementsFactory.NewApiEndpointRequirement())
	}

	if c.String("o") != "" || c.String("s") != "" {
		reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	}

	if !cmd.config.IsLoggedIn() && c.String("a") != "" && (c.String("o") != "" || c.String("s") != "") {
		err = errors.NewWithFmt(T("user not logged in, cannot use [-o] or [-s] flag to target"))

		cmd.ui.Failed(err.Error())
	}
	return
}

func (cmd Target) Run(c *cli.Context) {
	orgName := c.String("o")
	spaceName := c.String("s")
	apiEndpoint := c.String("a")

	if apiEndpoint != "" {
		skipSSL := c.Bool("skip-ssl-validation")

		cmd.ui.Say(T("Setting api endpoint to {{.Endpoint}}...",
			map[string]interface{}{"Endpoint": terminal.EntityNameColor(apiEndpoint)}))
		NewApi(cmd.ui, cmd.config, cmd.endpointRepo).setApiEndpoint(apiEndpoint, skipSSL, cmd.Metadata().Name)
	}

	if orgName != "" {
		err := cmd.setOrganization(orgName)
		if err != nil {
			cmd.ui.Failed(err.Error())
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
	return
}

func (cmd Target) setOrganization(orgName string) error {
	// setting an org necessarily invalidates any space you had previously targeted
	cmd.config.SetOrganizationFields(models.OrganizationFields{})
	cmd.config.SetSpaceFields(models.SpaceFields{})

	org, apiErr := cmd.orgRepo.FindByName(orgName)
	if apiErr != nil {
		return errors.NewWithFmt(T("Could not target org.\n{{.ApiErr}}",
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
		return errors.NewWithFmt(T("Unable to access space {{.SpaceName}}.\n{{.ApiErr}}",
			map[string]interface{}{"SpaceName": spaceName, "ApiErr": apiErr.Error()}))
	}

	cmd.config.SetSpaceFields(space.SpaceFields)
	return nil
}
