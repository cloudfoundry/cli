package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
)

type Target struct {
	ui        terminal.UI
	config    configuration.ReadWriter
	orgRepo   api.OrganizationRepository
	spaceRepo api.SpaceRepository
}

func NewTarget(ui terminal.UI,
	config configuration.ReadWriter,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (cmd Target) {

	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	cmd.spaceRepo = spaceRepo

	return
}

func (cmd Target) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 0 {
		err = errors.New("incorrect usage")
		cmd.ui.FailWithUsage(c, "target")
		return
	}

	if c.String("o") != "" || c.String("s") != "" {
		reqs = append(reqs, reqFactory.NewLoginRequirement())
	}
	return
}

func (cmd Target) Run(c *cli.Context) {
	orgName := c.String("o")
	spaceName := c.String("s")
	shouldShowTarget := (orgName == "" && spaceName == "")

	if shouldShowTarget {
		cmd.ui.ShowConfiguration(cmd.config)
		return
	}

	if orgName != "" {
		err := cmd.setOrganization(orgName)

		if spaceName == "" && cmd.config.IsLoggedIn() {
			cmd.ui.ShowConfiguration(cmd.config)
			return
		}

		if err != nil {
			return
		}
	}

	if spaceName != "" {
		err := cmd.setSpace(spaceName)

		if err != nil {
			return
		}
	}
	cmd.ui.ShowConfiguration(cmd.config)
	return
}

func (cmd Target) setOrganization(orgName string) (err error) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to target an org. Use '%s'.", terminal.CommandColor(cf.Name()+" login"))
		return
	}

	org, apiResponse := cmd.orgRepo.FindByName(orgName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Could not target org.\n%s", apiResponse.Message)
		return
	}

	cmd.config.SetOrganizationFields(org.OrganizationFields)
	return
}

func (cmd Target) setSpace(spaceName string) (err error) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set a space. Use '%s'.", terminal.CommandColor(fmt.Sprintf("%s login", cf.Name())))
		return
	}

	if !cmd.config.HasOrganization() {
		cmd.ui.Failed("An org must be targeted before targeting a space")
		return
	}

	space, apiResponse := cmd.spaceRepo.FindByName(spaceName)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Unable to access space %s.\n%s", spaceName, apiResponse.Message)
		return
	}

	cmd.config.SetSpaceFields(space.SpaceFields)
	return
}
