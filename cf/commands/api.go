package commands

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type API struct {
	ui           terminal.UI
	endpointRepo coreconfig.EndpointRepository
	config       coreconfig.ReadWriter
}

func init() {
	commandregistry.Register(API{})
}

func (cmd API) MetaData() commandregistry.CommandMetadata {
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

func (cmd API) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd API) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.endpointRepo = deps.RepoLocator.GetEndpointRepository()
	return cmd
}

func (cmd API) Execute(c flags.FlagContext) error {
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
		err := cmd.setAPIEndpoint(endpoint, c.Bool("skip-ssl-validation"), cmd.MetaData().Name)
		if err != nil {
			return err
		}
		cmd.ui.Ok()

		cmd.ui.Say("")
		cmd.ui.ShowConfiguration(cmd.config)
	}
	return nil
}

func (cmd API) setAPIEndpoint(endpoint string, skipSSL bool, cmdName string) error {
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
			return errors.New(T("Invalid SSL Cert for {{.URL}}\n{{.TipMessage}}",
				map[string]interface{}{"URL": typedErr.URL, "TipMessage": tipMessage}))
		default:
			return typedErr
		}
	}

	if warning != nil {
		cmd.ui.Say(terminal.WarningColor(warning.Warn()))
	}
	return nil
}
