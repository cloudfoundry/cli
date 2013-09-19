package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
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

	onConnect := func() {
		cmd.ui.Say("Connected, tailing...")
	}

	onMessage := func(lm *logmessage.LogMessage) {
		cmd.ui.Say(logMessageOutput(lm))
	}

	err := cmd.logsRepo.TailLogsFor(app, onConnect, onMessage)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}

func logMessageOutput(lm *logmessage.LogMessage) string {
	sourceType, _ := logmessage.LogMessage_SourceType_name[int32(*lm.SourceType)]
	sourceId := "?"
	if lm.SourceId != nil {
		sourceId = *lm.SourceId
	}
	msg := lm.GetMessage()

	return fmt.Sprintf("[%s/%s] %s", sourceType, sourceId, msg)
}
