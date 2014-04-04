package application

import (
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/codegangsta/cli"
	"time"
)

type Logs struct {
	ui       terminal.UI
	config   configuration.Reader
	logsRepo api.LogsRepository
	appReq   requirements.ApplicationRequirement
}

func NewLogs(ui terminal.UI, config configuration.Reader, logsRepo api.LogsRepository) (cmd *Logs) {
	cmd = new(Logs)
	cmd.ui = ui
	cmd.config = config
	cmd.logsRepo = logsRepo
	return
}

func (cmd *Logs) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c, "logs")
		err = errors.New("Incorrect Usage")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
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
	cmd.ui.Say("Connected, dumping recent logs for app %s in org %s / space %s as %s...\n",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	messages, err := cmd.logsRepo.RecentLogsFor(app.Guid)
	if err != nil {
		cmd.handleError(err)
	}

	for _, msg := range messages {
		cmd.ui.Say("%s", LogMessageOutput(msg, time.Local))
	}
}

func (cmd *Logs) tailLogsFor(app models.Application) {
	onConnect := func() {
		cmd.ui.Say("Connected, tailing logs for app %s in org %s / space %s as %s...\n",
			terminal.EntityNameColor(app.Name),
			terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)
	}

	err := cmd.logsRepo.TailLogsFor(app.Guid, 5*time.Second, onConnect, func(msg *logmessage.LogMessage) {
		cmd.ui.Say("%s", LogMessageOutput(msg, time.Local))
	})

	if err != nil {
		cmd.handleError(err)
	}
}

func (cmd *Logs) handleError(err error) {
	switch err.(type) {
	case nil:
	case *errors.InvalidSSLCert:
		cmd.ui.Failed(err.Error() + "\nTIP: use 'cf login -a API --skip-ssl-validation' or 'cf api API --skip-ssl-validation' to suppress this error")
	default:
		cmd.ui.Failed(err.Error())
	}
}
