package serviceauthtoken

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListServiceAuthTokens struct {
	ui            terminal.UI
	config        coreconfig.Reader
	authTokenRepo api.ServiceAuthTokenRepository
}

func init() {
	commandregistry.Register(&ListServiceAuthTokens{})
}

func (cmd *ListServiceAuthTokens) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "service-auth-tokens",
		Description: T("List service auth tokens"),
		Usage: []string{
			T("CF_NAME service-auth-tokens"),
		},
	}
}

func (cmd *ListServiceAuthTokens) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewMaxAPIVersionRequirement(
			"service-auth-tokens",
			cf.ServiceAuthTokenMaximumAPIVersion,
		),
	}

	return reqs
}

func (cmd *ListServiceAuthTokens) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.authTokenRepo = deps.RepoLocator.GetServiceAuthTokenRepository()
	return cmd
}

func (cmd *ListServiceAuthTokens) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Getting service auth tokens as {{.CurrentUser}}...",
		map[string]interface{}{
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))
	authTokens, err := cmd.authTokenRepo.FindAll()
	if err != nil {
		return err
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	table := cmd.ui.Table([]string{T("label"), T("provider")})

	for _, authToken := range authTokens {
		table.Add(authToken.Label, authToken.Provider)
	}

	table.Print()
	return nil
}
