package application

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/cf/api/logs"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
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
	fs["recent"] = &flags.BoolFlag{Name: "recent", Usage: T("Dump recent logs instead of tailing")}

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

	var err error
	if c.Bool("recent") {
		err = cmd.recentLogsFor(app)
	} else {
		err = cmd.tailLogsFor(app)
	}
	if err != nil {
		return err
	}
	return nil
}

func (cmd *Logs) recentLogsFor(app models.Application) error {
	cmd.ui.Say(T("Connected, dumping recent logs for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	messages, err := cmd.logsRepo.RecentLogsFor(app.GUID)
	if err != nil {
		return cmd.handleError(err)
	}

	for _, msg := range messages {
		cmd.ui.Say("%s", msg.ToLog(time.Local))
	}
	return nil
}

func (cmd *Logs) tailLogsFor(app models.Application) error {
	onConnect := func() {
		cmd.ui.Say(T("Connected, tailing logs for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...\n",
			map[string]interface{}{
				"AppName":   terminal.EntityNameColor(app.Name),
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":  terminal.EntityNameColor(cmd.config.Username())}))
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
			cmd.ui.Say("%s", msg.ToLog(time.Local))
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
