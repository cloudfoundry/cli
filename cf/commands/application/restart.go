package application

import (
	"os"

	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Restart struct {
	ui           terminal.UI
	config       core_config.Reader
	starter      ApplicationStarter
	stopper      ApplicationStopper
	appRepo      applications.ApplicationRepository
	manifestRepo manifest.ManifestRepository
}

type ApplicationRestarter interface {
	ApplicationRestart(app models.Application, orgName string, spaceName string)
}

func NewRestart(ui terminal.UI, config core_config.Reader, starter ApplicationStarter, stopper ApplicationStopper, manifestRepo manifest.ManifestRepository, appRepo applications.ApplicationRepository) (cmd *Restart) {
	cmd = new(Restart)
	cmd.ui = ui
	cmd.config = config
	cmd.starter = starter
	cmd.stopper = stopper
	cmd.manifestRepo = manifestRepo
	cmd.appRepo = appRepo
	return
}

func (cmd *Restart) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "restart",
		ShortName:   "rs",
		Description: T("Restart an app"),
		Usage: T("Restart a single app:\n") + T("   CF_NAME restart APP_NAME\n") +
			"\n" + T("   Restart an app(s) with a manifest from a current directory:\n") + T("   CF_NAME restart\n"),
	}
}

func (cmd *Restart) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) > 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *Restart) Run(c *cli.Context) {
	appNames := cmd.findAppNamesToRestart(c)
	for _, appName := range appNames {
		app, apiErr := cmd.appRepo.Read(appName)
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
			return
		}
		cmd.ApplicationRestart(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	}
}

func (cmd *Restart) ApplicationRestart(app models.Application, orgName, spaceName string) {
	stoppedApp, err := cmd.stopper.ApplicationStop(app, orgName, spaceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Say("")

	_, err = cmd.starter.ApplicationStart(stoppedApp, orgName, spaceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}

func (cmd *Restart) findAppNamesToRestart(c *cli.Context) []string {
	if len(c.Args()) > 0 {
		return []string{c.Args()[0]}
	} else {
		return cmd.findAppsFromManifest(c)
	}
}

func (cmd *Restart) findAppsFromManifest(c *cli.Context) []string {
	var err error
	var path string

	path, err = os.Getwd()
	if err != nil {
		cmd.ui.Failed(T("Could not determine the current working directory!"), err)
	}

	m, err := cmd.manifestRepo.ReadManifest(path)
	if err != nil {
		cmd.ui.Warn(T("Error reading manifest file:\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
		cmd.ui.Say("")
		cmd.ui.FailWithUsage(c)
	}

	apps, err := m.Applications()
	if err != nil {
		cmd.ui.Failed("Error reading manifest file:\n%s", err)
	}
	cmd.ui.Say(T("Using manifest file {{.Path}}\n",
		map[string]interface{}{"Path": terminal.EntityNameColor(m.Path)}))

	appNames := make([]string, len(apps))
	for index, app := range apps {
		appNames[index] = *app.Name
	}

	return appNames
}
