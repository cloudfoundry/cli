package application

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/noaa/events"
	"github.com/codegangsta/cli"
)

type Logs struct {
	ui          terminal.UI
	config      core_config.Reader
	noaaRepo    api.LogsNoaaRepository
	oldLogsRepo api.OldLogsRepository
	appReq      requirements.ApplicationRequirement
}

func NewLogs(ui terminal.UI, config core_config.Reader, noaaRepo api.LogsNoaaRepository, oldLogsRepo api.OldLogsRepository) (cmd *Logs) {
	cmd = new(Logs)
	cmd.ui = ui
	cmd.config = config
	cmd.noaaRepo = noaaRepo
	cmd.oldLogsRepo = oldLogsRepo
	return
}

func (cmd *Logs) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "logs",
		Description: T("Tail or show recent logs for an app"),
		Usage:       T("CF_NAME logs APP_NAME"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "recent", Usage: T("Dump recent logs instead of tailing")},
		},
	}
}

func (cmd *Logs) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	if cmd.appReq == nil {
		cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	} else {
		cmd.appReq.SetApplicationName(c.Args()[0])
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *Logs) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	if c.Bool("recent") {
		cmd.recentLogsFor(app)
	} else {
		cmd.tailLogsFor(app)
	}
}

func (cmd *Logs) recentLogsFor(app models.Application) {
	cmd.ui.Say(T("Connected, dumping recent logs for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	messages, err := cmd.oldLogsRepo.RecentLogsFor(app.Guid)
	// messages, err := cmd.noaaRepo.RecentLogsFor(app.Guid)
	if err != nil {
		cmd.handleError(err)
	}

	for _, msg := range messages {
		cmd.ui.Say("%s", LogMessageOutput(msg, time.Local))
		// cmd.ui.Say("%s", LogNoaaMessageOutput(msg, time.Local))
	}
}

func (cmd *Logs) tailLogsFor(app models.Application) {
	onConnect := func() {
		cmd.ui.Say(T("Connected, tailing logs for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
			map[string]interface{}{
				"AppName":   terminal.EntityNameColor(app.Name),
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":  terminal.EntityNameColor(cmd.config.Username())}))
	}

	err := cmd.oldLogsRepo.TailLogsFor(app.Guid, onConnect, func(msg *logmessage.LogMessage) {
		cmd.ui.Say("%s", LogMessageOutput(msg, time.Local))
	})
	// err := cmd.noaaRepo.TailNoaaLogsFor(app.Guid, onConnect, func(msg *events.LogMessage) {
	// 	cmd.ui.Say("%s", LogNoaaMessageOutput(msg, time.Local))
	// })

	if err != nil {
		cmd.handleError(err)
	}
}

func (cmd *Logs) handleError(err error) {
	switch err.(type) {
	case nil:
	case *errors.InvalidSSLCert:
		cmd.ui.Failed(err.Error() + T("\nTIP: use 'cf login -a API --skip-ssl-validation' or 'cf api API --skip-ssl-validation' to suppress this error"))
	default:
		cmd.ui.Failed(err.Error())
	}
}

func LogMessageOutput(msg *logmessage.LogMessage, loc *time.Location) string {
	logHeader, coloredLogHeader := ui_helpers.ExtractLogHeader(msg, loc)
	logContent := ui_helpers.ExtractLogContent(msg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}

func LogNoaaMessageOutput(msg *events.LogMessage, loc *time.Location) string {
	logHeader, coloredLogHeader := ui_helpers.ExtractNoaaLogHeader(msg, loc)
	logContent := ui_helpers.ExtractNoaaLogContent(msg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}
