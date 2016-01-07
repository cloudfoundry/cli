package commands

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
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

var ErrNoRedirects = errors.New("No redirects")

func init() {
	command_registry.Register(OneTimeSSHCode{})
}

func (cmd OneTimeSSHCode) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "ssh-code",
		Description: T("Get a one time password for ssh clients"),
		Usage:       T("CF_NAME ssh-code"),
	}
}

func (cmd OneTimeSSHCode) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("ssh-code"))
	}

	reqs := append([]requirements.Requirement{}, requirementsFactory.NewApiEndpointRequirement())
	return reqs, nil
}

func (cmd OneTimeSSHCode) SetDependency(deps command_registry.Dependency, _ bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.endpointRepo = deps.RepoLocator.GetEndpointRepository()

	return cmd
}

func (cmd OneTimeSSHCode) Execute(c flags.FlagContext) {
	code, err := cmd.Get()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(code)
}

func (cmd OneTimeSSHCode) Get() (string, error) {
	_, err := cmd.endpointRepo.UpdateEndpoint(cmd.config.ApiEndpoint())
	if err != nil {
		return "", errors.New(T("Error getting info from v2/info: ") + err.Error())
	}

	token, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		return "", errors.New(T("Error refreshing oauth token: ") + err.Error())
	}

	skipCertificateVerify := cmd.config.IsSSLDisabled()

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			dumpRequest(req)
			return ErrNoRedirects
		},
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipCertificateVerify,
			},
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	authorizeURL, err := cmd.authorizeURL()
	if err != nil {
		return "", errors.New(T("Error getting AuthenticationEndpoint() ") + err.Error())
	}

	authorizeReq, err := http.NewRequest("GET", authorizeURL, nil)
	if err != nil {
		return "", err
	}

	authorizeReq.Header.Add("authorization", token)

	resp, err := httpClient.Do(authorizeReq)
	if resp != nil {
		dumpResponse(resp)
	}
	if err == nil {
		return "", errors.New(T("Authorization server did not redirect with one time code"))
	}

	if netErr, ok := err.(*url.Error); !ok || netErr.Err != ErrNoRedirects {
		return "", errors.New(T("Error requesting one time code from server:" + err.Error()))
	}

	loc, err := resp.Location()
	if err != nil {
		return "", errors.New(T("Error getting the redirected lcoation: " + err.Error()))
	}

	codes := loc.Query()["code"]
	if len(codes) != 1 {
		return "", errors.New(T("Unable to acquire one time code from authorization response") + "\n" + T("Server did not response with auth code") + "\n")
	}

	return codes[0], nil
}

func (cmd OneTimeSSHCode) authorizeURL() (string, error) {
	authorizeURL, err := url.Parse(cmd.config.UaaEndpoint())
	if err != nil {
		return "", err
	}

	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", cmd.config.SSHOAuthClient())

	authorizeURL.Path = "/oauth/authorize"
	authorizeURL.RawQuery = values.Encode()

	return authorizeURL.String(), nil
}

func dumpRequest(req *http.Request) {
	shouldDisplayBody := !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data")
	dumpedRequest, err := httputil.DumpRequestOut(req, shouldDisplayBody)
	if err != nil {
		trace.Logger.Printf(T("Error dumping request\n{{.Err}}\n", map[string]interface{}{"Err": err}))
	} else {
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor(T("REQUEST:")), time.Now().Format(time.RFC3339), trace.Sanitize(string(dumpedRequest)))
		if !shouldDisplayBody {
			trace.Logger.Println(T("[MULTIPART/FORM-DATA CONTENT HIDDEN]"))
		}
	}
}

func dumpResponse(res *http.Response) {
	dumpedResponse, err := httputil.DumpResponse(res, false)
	if err != nil {
		trace.Logger.Printf(T("Error dumping response\n{{.Err}}\n", map[string]interface{}{"Err": err}))
	} else {
		trace.Logger.Printf("\n%s [%s]\n%s\n", terminal.HeaderColor(T("RESPONSE:")), time.Now().Format(time.RFC3339), trace.Sanitize(string(dumpedResponse)))
	}
}
