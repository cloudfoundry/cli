package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Files struct {
	ui           terminal.UI
	config       configuration.Reader
	appFilesRepo api.AppFilesRepository
	appReq       requirements.ApplicationRequirement
}

func NewFiles(ui terminal.UI, config configuration.Reader, appFilesRepo api.AppFilesRepository) (cmd *Files) {
	cmd = new(Files)
	cmd.ui = ui
	cmd.config = config
	cmd.appFilesRepo = appFilesRepo
	return
}

func (cmd *Files) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "files",
		ShortName:   "f",
		Description: T("Print out a list of files in a directory or the contents of a specific file"),
		Usage:       T("CF_NAME files APP [PATH]"),
	}
}

func (cmd *Files) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 1 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Files) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	cmd.ui.Say(T("Getting files for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	path := "/"
	if len(c.Args()) > 1 {
		path = c.Args()[1]
	}

	list, apiErr := cmd.appFilesRepo.ListFiles(app.Guid, path)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
	cmd.ui.Say("%s", list)
}
