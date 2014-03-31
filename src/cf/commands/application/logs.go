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
	logChan := make(chan *logmessage.Message, 1000)
	errChan := make(chan error)

	go func() {
		defer close(logChan)
		if c.Bool("recent") {
			cmd.recentLogsFor(app, logChan, errChan)
		} else {
			cmd.tailLogsFor(app, logChan, errChan)
		}
	}()

	cmd.displayLogMessages(logChan, errChan)
}

func (cmd *Logs) recentLogsFor(app models.Application, logChan chan *logmessage.Message, errChan chan error) {
	onConnect := func() {
		cmd.ui.Say("Connected, dumping recent logs for app %s in org %s / space %s as %s...\n",
			terminal.EntityNameColor(app.Name),
			terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)
	}

	err := cmd.logsRepo.RecentLogsFor(app.Guid, onConnect, logChan)
	if err != nil {
		errChan <- err
	}
}

func (cmd *Logs) tailLogsFor(app models.Application, logChan chan *logmessage.Message, errChan chan error) {
	onConnect := func() {
		cmd.ui.Say("Connected, tailing logs for app %s in org %s / space %s as %s...\n",
			terminal.EntityNameColor(app.Name),
			terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)
	}

	// in this case we tail the logs forever, so we never send true on this channel
	stopLoggingChan := make(chan bool)
	defer close(stopLoggingChan)

	err := cmd.logsRepo.TailLogsFor(app.Guid, onConnect, logChan, stopLoggingChan, 5*time.Second)
	if err != nil {
		errChan <- err
	}
}

func (cmd *Logs) displayLogMessages(logChan chan *logmessage.Message, errChan chan error) {
	for {
		select {
		case err := <-errChan:
			switch err.(type) {
			case nil:
			case *errors.InvalidSSLCert:
				cmd.ui.Failed(err.Error() + "\nTIP: use 'cf login -a API --skip-ssl-validation' or 'cf api API --skip-ssl-validation' to suppress this error")
			default:
				cmd.ui.Failed(err.Error())
			}

		case msg, ok := <-logChan:
			if !ok {
				return
			}
			cmd.ui.Say("%s", LogMessageOutput(msg))
		}
	}
}
