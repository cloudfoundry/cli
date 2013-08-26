package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type Apps struct {
	ui      term.UI
	appRepo api.ApplicationRepository
	config  *configuration.Configuration
}

func NewApps(ui term.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (a Apps) {
	a.ui = ui
	a.appRepo = appRepo
	a.config = config
	return
}

func (a Apps) Run(c *cli.Context) {
	a.ui.Say("Getting applications in %s", a.config.Space.Name)

	apps, err := a.appRepo.FindAll(a.config)

	if err != nil {
		a.ui.Failed("Error loading applications", err)
		return
	}

	a.ui.Ok()

	longestNameLength := len(longestName(apps))
	appNamePadding := strings.Repeat(" ", longestNameLength-len("name"))
	a.ui.Say("name%s \tstatus  \tusage   \turl", appNamePadding)

	for _, app := range apps {
		appNamePadding := strings.Repeat(" ", longestNameLength-len(app.Name))

		a.ui.Say(
			"%s%s \t%s \t%d x %dM \t%s",
			term.Cyan(app.Name),
			appNamePadding,
			coloredState(app.State),
			app.Instances,
			app.Memory,
			strings.Join(app.Urls, ", "),
		)
	}
}

func longestName(apps []cf.Application) (name string) {
	for _, app := range apps {
		if len(app.Name) > len(name) {
			name = app.Name
		}
	}

	return
}

func coloredState(state string) (colored string) {
	switch state {
	case "started":
		colored = term.Green("running")
	case "stopped":
		colored = term.Yellow("stopped")
	default:
		colored = term.Red(state)
	}

	return
}
