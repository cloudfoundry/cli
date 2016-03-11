package commands

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter -o fakes/fake_ssh_code_getter.go . SSHCodeGetter
type SSHCodeGetter interface {
	command_registry.Command
	Get() (string, error)
}

type OneTimeSSHCode struct {
	ui           terminal.UI
	config       core_config.ReadWriter
	authRepo     authentication.AuthenticationRepository
	endpointRepo api.EndpointRepository
}

func init() {
	command_registry.Register(&OneTimeSSHCode{})
}

func (cmd *OneTimeSSHCode) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "ssh-code",
		Description: T("Get a one time password for ssh clients"),
		Usage: []string{
			T("CF_NAME ssh-code"),
		},
	}
}

func (cmd *OneTimeSSHCode) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(command_registry.CliCommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewApiEndpointRequirement(),
	}

	return reqs
}

func (cmd *OneTimeSSHCode) SetDependency(deps command_registry.Dependency, _ bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.endpointRepo = deps.RepoLocator.GetEndpointRepository()

	return cmd
}

func (cmd *OneTimeSSHCode) Execute(c flags.FlagContext) {
	code, err := cmd.Get()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(code)
}

func (cmd *OneTimeSSHCode) Get() (string, error) {
	_, err := cmd.endpointRepo.UpdateEndpoint(cmd.config.ApiEndpoint())
	if err != nil {
		return "", errors.New(T("Error getting info from v2/info: ") + err.Error())
	}

	token, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		return "", errors.New(T("Error refreshing oauth token: ") + err.Error())
	}

	sshCode, err := cmd.authRepo.Authorize(token)
	if err != nil {
		return "", errors.New(T("Error getting SSH code: ") + err.Error())
	}

	return sshCode, nil
}
