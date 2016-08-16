package application

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api/logs"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type Logs struct {
	ui       terminal.UI
	logsRepo logs.Repository
	config   coreconfig.Reader
	appReq   requirements.ApplicationRequirement
}

func init() {
	commandregistry.Register(&Logs{})
}

func (cmd *Logs) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["recent"] = &flags.BoolFlag{
		Name:  "recent",
		Usage: T("Dump recent logs instead of tailing"),
	}
	fs["newline"] = &flags.StringFlag{
		Name:  "newline",
		Usage: T("Replace the specified rune (in hex) with a newline"),
	}

	return commandregistry.CommandMetadata{
		Name:        "logs",
		Description: T("Tail or show recent logs for an app"),
		Usage: []string{
			T("CF_NAME logs APP_NAME"),
		},
		Flags: fs,
	}
}

func (cmd *Logs) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("logs"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *Logs) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.logsRepo = deps.RepoLocator.GetLogsRepository()
	return cmd
}

func (cmd *Logs) Execute(c flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	r := rune(-1)

	sr := c.String("newline")
	if sr != "" {
		cp, err := strconv.ParseInt(sr, 16, 32)
		if err != nil {
			return err
		}
		r = rune(cp)
	}

	if c.Bool("recent") {
		return cmd.recentLogsFor(app, r)
	}
	return cmd.tailLogsFor(app, r)
}

func (cmd *Logs) recentLogsFor(app models.Application, newline rune) error {
	cmd.ui.Say(T("Connected, dumping recent logs for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username()),
		}))

	messages, err := cmd.logsRepo.RecentLogsFor(app.GUID)
	if err != nil {
		return cmd.handleError(err)
	}

	for _, msg := range messages {
		cmd.ui.Say("%s", newlineReplacement(newline, msg.ToLog(time.Local)))
	}
	return nil
}

func (cmd *Logs) tailLogsFor(app models.Application, newline rune) error {
	onConnect := func() {
		cmd.ui.Say(T("Connected, tailing logs for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
			map[string]interface{}{
				"AppName":   terminal.EntityNameColor(app.Name),
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":  terminal.EntityNameColor(cmd.config.Username()),
			}))
	}

	c := make(chan logs.Loggable)
	e := make(chan error)

	go cmd.logsRepo.TailLogsFor(app.GUID, onConnect, c, e)

	for {
		select {
		case msg, ok := <-c:
			if !ok {
				return nil
			}
			cmd.ui.Say("%s", newlineReplacement(newline, msg.ToLog(time.Local)))
		case err := <-e:
			return cmd.handleError(err)
		}
	}
}

func (cmd *Logs) handleError(err error) error {
	switch err.(type) {
	case nil:
	case *errors.InvalidSSLCert:
		return errors.New(err.Error() + T("\nTIP: use 'cf login -a API --skip-ssl-validation' or 'cf api API --skip-ssl-validation' to suppress this error"))
	default:
		return err
	}
	return nil
}

func newlineReplacement(newline rune, msg string) string {
	return strings.Map(func(r rune) rune {
		if r == newline {
			return '\n'
		}
		return r
	}, msg)
}
