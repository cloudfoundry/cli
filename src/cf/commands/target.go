package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"os"
)

type Target struct {
	ui         terminal.UI
	config     *configuration.Configuration
	configRepo configuration.ConfigurationRepository
	orgRepo    api.OrganizationRepository
	spaceRepo  api.SpaceRepository
}

func NewTarget(ui terminal.UI, configRepo configuration.ConfigurationRepository, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository) (cmd Target) {
	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.config, _ = configRepo.Get()
	cmd.orgRepo = orgRepo
	cmd.spaceRepo = spaceRepo

	return
}

func (cmd Target) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) > 0 {
		err = errors.New("incorrect usage")
		cmd.ui.Say(terminal.FailureColor("FAILED"))
		cmd.ui.Say("Incorrect Usage.\n")
		cli.ShowCommandHelp(c, "target")
		cmd.ui.Say("")
		cmd.ui.Say("TIP:\n  Use 'cf api' to set or view the target api url.\n")
		os.Exit(1)

		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd Target) Run(c *cli.Context) {
	orgName := c.String("o")
	spaceName := c.String("s")

	if orgName == "" && spaceName == "" {
		cmd.ui.ShowConfiguration(cmd.config)

		if !cmd.config.HasOrganization() {
			cmd.ui.Say("No org targeted, use '%s target -o' to target an org", cf.Name)
		}
		if !cmd.config.HasSpace() {
			cmd.ui.Say("No space targeted, use '%s target -s' to target a space", cf.Name)
		}
		return
	}

	if orgName != "" {
		err := cmd.setOrganization(orgName)

		if spaceName == "" && cmd.config.IsLoggedIn() {
			cmd.showConfig()
			cmd.ui.Say("No space targeted, use '%s target -s' to target a space", cf.Name)
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
	cmd.showConfig()
	return
}

func (cmd Target) setOrganization(orgName string) (err error) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to target an org. Use '%s login'.", cf.Name)
		return
	}

	org, apiStatus := cmd.orgRepo.FindByName(orgName)
	if apiStatus.NotSuccessful() {
		cmd.ui.Failed("Could not target org.\n%s", apiStatus.Message)
		return
	}

	cmd.config.Organization = org
	cmd.config.Space = cf.Space{}
	cmd.saveConfig()
	return
}

func (cmd Target) setSpace(spaceName string) (err error) {
	if !cmd.config.IsLoggedIn() {
		cmd.ui.Failed("You must be logged in to set a space. Use '%s login'.", cf.Name)
		return
	}

	if !cmd.config.HasOrganization() {
		cmd.ui.Failed("An org must be targeted before targeting a space")
		return
	}

	space, apiStatus := cmd.spaceRepo.FindByName(spaceName)

	if apiStatus.NotSuccessful() {
		cmd.ui.Failed("Unable to access space %s.\n%s", spaceName, apiStatus.Message)
		return
	}

	cmd.config.Space = space
	cmd.saveConfig()
	return
}

func (cmd Target) saveConfig() {
	err := cmd.configRepo.Save()
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}

func (cmd Target) showConfig() {
	cmd.ui.ShowConfiguration(cmd.config)
}
