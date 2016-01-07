package commands

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type Api struct {
	ui           terminal.UI
	endpointRepo api.EndpointRepository
	config       core_config.ReadWriter
}

func init() {
	command_registry.Register(Api{})
}

func (cmd Api) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["unset"] = &cliFlags.BoolFlag{Name: "unset", Usage: T("Remove all api endpoint targeting")}
	fs["skip-ssl-validation"] = &cliFlags.BoolFlag{Name: "skip-ssl-validation", Usage: T("Please don't")}

	return command_registry.CommandMetadata{
		Name:        "api",
		Description: T("Set or view target api url"),
		Usage:       T("CF_NAME api [URL]"),
		Flags:       fs,
	}
}

func (cmd Api) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Api) SetDependency(deps command_registry.Dependency, _ bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.endpointRepo = deps.RepoLocator.GetEndpointRepository()
	return cmd
}

func (cmd Api) Execute(c flags.FlagContext) {
	if c.Bool("unset") {
		cmd.ui.Say(T("Unsetting api endpoint..."))
		cmd.config.SetApiEndpoint("")

		cmd.ui.Ok()
		cmd.ui.Say(T("\nNo api endpoint set."))

	} else if len(c.Args()) == 0 {
		if cmd.config.ApiEndpoint() == "" {
			cmd.ui.Say(fmt.Sprintf(T("No api endpoint set. Use '{{.Name}}' to set an endpoint",
				map[string]interface{}{"Name": terminal.CommandColor(cf.Name() + " api")})))
		} else {
			cmd.ui.Say(T("API endpoint: {{.ApiEndpoint}} (API version: {{.ApiVersion}})",
				map[string]interface{}{"ApiEndpoint": terminal.EntityNameColor(cmd.config.ApiEndpoint()),
					"ApiVersion": terminal.EntityNameColor(cmd.config.ApiVersion())}))
		}
	} else {
		endpoint := c.Args()[0]

		cmd.ui.Say(T("Setting api endpoint to {{.Endpoint}}...",
			map[string]interface{}{"Endpoint": terminal.EntityNameColor(endpoint)}))
		cmd.setApiEndpoint(endpoint, c.Bool("skip-ssl-validation"), cmd.MetaData().Name)
		cmd.ui.Ok()

		cmd.ui.Say("")
		cmd.ui.ShowConfiguration(cmd.config)
	}
}

func (cmd Api) setApiEndpoint(endpoint string, skipSSL bool, cmdName string) {
	if strings.HasSuffix(endpoint, "/") {
		endpoint = strings.TrimSuffix(endpoint, "/")
	}

	cmd.config.SetSSLDisabled(skipSSL)
	endpoint, err := cmd.endpointRepo.UpdateEndpoint(endpoint)

	if err != nil {
		cmd.config.SetApiEndpoint("")
		cmd.config.SetSSLDisabled(false)

		switch typedErr := err.(type) {
		case *errors.InvalidSSLCert:
			cfApiCommand := terminal.CommandColor(fmt.Sprintf("%s %s --skip-ssl-validation", cf.Name(), cmdName))
			tipMessage := fmt.Sprintf(T("TIP: Use '{{.ApiCommand}}' to continue with an insecure API endpoint",
				map[string]interface{}{"ApiCommand": cfApiCommand}))
			cmd.ui.Failed(T("Invalid SSL Cert for {{.URL}}\n{{.TipMessage}}",
				map[string]interface{}{"URL": typedErr.URL, "TipMessage": tipMessage}))
		default:
			cmd.ui.Failed(typedErr.Error())
		}
	}

	if !strings.HasPrefix(endpoint, "https://") {
		cmd.ui.Say(terminal.WarningColor(T("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended\n")))
	}
}
