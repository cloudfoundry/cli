package application

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/codegangsta/cli"
)

type Logs struct {
	ui       terminal.UI
	logsRepo api.LogsRepository
	appReq   requirements.ApplicationRequirement
}

func NewLogs(ui terminal.UI, logsRepo api.LogsRepository) (cmd *Logs) {
	cmd = new(Logs)
	cmd.ui = ui
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

	onMessage := func(msg logmessage.LogMessage) {
		cmd.ui.Say(logMessageOutput(app.Name, msg))
	}

	var err error

	if c.Bool("recent") {
		onConnect := func() {
			cmd.ui.Say("Connected, dumping recent logs...")
		}

		err = cmd.logsRepo.RecentLogsFor(app, onConnect, onMessage, "4443")
	} else {
		onConnect := func() {
			cmd.ui.Say("Connected, tailing...")
		}

		err = cmd.logsRepo.TailLogsFor(app, onConnect, onMessage, 2, "4443")
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
}
