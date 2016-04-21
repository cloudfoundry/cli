package commands

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Api struct {
	ui           terminal.UI
	endpointRepo coreconfig.EndpointRepository
	config       coreconfig.ReadWriter
}

func init() {
	commandregistry.Register(Api{})
}

func (cmd Api) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["unset"] = &flags.BoolFlag{Name: "unset", Usage: T("Remove all api endpoint targeting")}
	fs["skip-ssl-validation"] = &flags.BoolFlag{Name: "skip-ssl-validation", Usage: T("Skip verification of the API endpoint. Not recommended!")}

	return commandregistry.CommandMetadata{
		Name:        "api",
		Description: T("Set or view target api url"),
		Usage: []string{
			T("CF_NAME api [URL]"),
		},
		Flags: fs,
	}
}

func (cmd Api) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd Api) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.endpointRepo = deps.RepoLocator.GetEndpointRepository()
	return cmd
}

func (cmd Api) Execute(c flags.FlagContext) {
	if c.Bool("unset") {
		cmd.ui.Say(T("Unsetting api endpoint..."))
		cmd.config.SetAPIEndpoint("")

		cmd.ui.Ok()
		cmd.ui.Say(T("\nNo api endpoint set."))

	} else if len(c.Args()) == 0 {
		if cmd.config.APIEndpoint() == "" {
			cmd.ui.Say(fmt.Sprintf(T("No api endpoint set. Use '{{.Name}}' to set an endpoint",
				map[string]interface{}{"Name": terminal.CommandColor(cf.Name + " api")})))
		} else {
			cmd.ui.Say(T("API endpoint: {{.APIEndpoint}} (API version: {{.APIVersion}})",
				map[string]interface{}{"APIEndpoint": terminal.EntityNameColor(cmd.config.APIEndpoint()),
					"APIVersion": terminal.EntityNameColor(cmd.config.APIVersion())}))
		}
	} else {
		endpoint := c.Args()[0]

		cmd.ui.Say(T("Setting api endpoint to {{.Endpoint}}...",
			map[string]interface{}{"Endpoint": terminal.EntityNameColor(endpoint)}))
		cmd.setAPIEndpoint(endpoint, c.Bool("skip-ssl-validation"), cmd.MetaData().Name)
		cmd.ui.Ok()

		cmd.ui.Say("")
		cmd.ui.ShowConfiguration(cmd.config)
	}
}

func (cmd Api) setAPIEndpoint(endpoint string, skipSSL bool, cmdName string) {
	if strings.HasSuffix(endpoint, "/") {
		endpoint = strings.TrimSuffix(endpoint, "/")
	}

	cmd.config.SetSSLDisabled(skipSSL)

	refresher := coreconfig.APIConfigRefresher{
		Endpoint:     endpoint,
		EndpointRepo: cmd.endpointRepo,
		Config:       cmd.config,
	}

	warning, err := refresher.Refresh()
	if err != nil {
		cmd.config.SetAPIEndpoint("")
		cmd.config.SetSSLDisabled(false)

		switch typedErr := err.(type) {
		case *errors.InvalidSSLCert:
			cfAPICommand := terminal.CommandColor(fmt.Sprintf("%s %s --skip-ssl-validation", cf.Name, cmdName))
			tipMessage := fmt.Sprintf(T("TIP: Use '{{.APICommand}}' to continue with an insecure API endpoint",
				map[string]interface{}{"APICommand": cfAPICommand}))
			cmd.ui.Failed(T("Invalid SSL Cert for {{.URL}}\n{{.TipMessage}}",
				map[string]interface{}{"URL": typedErr.URL, "TipMessage": tipMessage}))
			return
		default:
			cmd.ui.Failed(typedErr.Error())
			return
		}
	}

	if warning != nil {
		cmd.ui.Say(terminal.WarningColor(warning.Warn()))
	}
}
