package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/codegangsta/cli"
)

type RecentLogs struct {
	ui       terminal.UI
	appRepo  api.ApplicationRepository
	appReq   requirements.ApplicationRequirement
	logsRepo api.LogsRepository
}

func NewRecentLogs(ui terminal.UI, aR api.ApplicationRepository, lR api.LogsRepository) (l *RecentLogs) {
	l = new(RecentLogs)
	l.ui = ui
	l.appRepo = aR
	l.logsRepo = lR
	return
}

func (l *RecentLogs) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		l.ui.FailWithUsage(c, "logs")
		return
	}

	l.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{l.appReq}
	return
}

func (l *RecentLogs) Run(c *cli.Context) {
	app := l.appReq.GetApplication()

	onConnect := func() {
		l.ui.Say("Connected, dumping recent logs...")
	}

	onMessage := func(msg logmessage.LogMessage) {
		l.ui.Say(logMessageOutput(app.Name, msg))
	}

	err := l.logsRepo.RecentLogsFor(app, onConnect, onMessage)
	if err != nil {
		l.ui.Failed(err.Error())
	}
}
