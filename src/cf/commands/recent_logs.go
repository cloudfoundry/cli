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

func NewRecentLogs(ui terminal.UI, aR api.ApplicationRepository, lR api.LogsRepository) (cmd *RecentLogs) {
	cmd = new(RecentLogs)
	cmd.ui = ui
	cmd.appRepo = aR
	cmd.logsRepo = lR
	return
}

func (cmd *RecentLogs) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "logs")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{cmd.appReq}
	return
}

func (cmd *RecentLogs) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()

	onConnect := func() {
		cmd.ui.Say("Connected, dumping recent logs...")
	}

	onMessage := func(msg logmessage.LogMessage) {
		cmd.ui.Say(logMessageOutput(app.Name, msg))
	}

	err := cmd.logsRepo.RecentLogsFor(app, onConnect, onMessage)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}
