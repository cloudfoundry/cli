package pluginrepo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"

	clipr "github.com/cloudfoundry/cli-plugin-repo/web"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type AddPluginRepo struct {
	ui     terminal.UI
	config coreconfig.ReadWriter
}

func init() {
	commandregistry.Register(&AddPluginRepo{})
}

func (cmd *AddPluginRepo) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "add-plugin-repo",
		Description: T("Add a new plugin repository"),
		Usage: []string{
			T(`CF_NAME add-plugin-repo REPO_NAME URL`),
		},
		Examples: []string{
			"CF_NAME add-plugin-repo PrivateRepo https://myprivaterepo.com/repo/",
		},
		TotalArgs: 2,
	}
}

func (cmd *AddPluginRepo) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires REPO_NAME and URL as arguments\n\n") + commandregistry.Commands.CommandUsage("add-plugin-repo"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd *AddPluginRepo) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	return cmd
}

func (cmd *AddPluginRepo) Execute(c flags.FlagContext) error {
	cmd.ui.Say("")
	repoURL := strings.ToLower(c.Args()[1])
	repoName := strings.Trim(c.Args()[0], " ")

	err := cmd.checkIfRepoExists(repoName, repoURL)
	if err != nil {
		return err
	}

	repoURL, err = cmd.verifyURL(repoURL)
	if err != nil {
		return err
	}

	resp, err := http.Get(repoURL)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if opErr, opErrOk := urlErr.Err.(*net.OpError); opErrOk {
				if opErr.Op == "dial" {
					return errors.New(T("There is an error performing request on '{{.RepoURL}}': {{.Error}}\n{{.Tip}}", map[string]interface{}{
						"RepoURL": repoURL,
						"Error":   err.Error(),
						"Tip":     T("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."),
					}))
				}
			}
		}
		return errors.New(T("There is an error performing request on '{{.RepoURL}}': ", map[string]interface{}{
			"RepoURL": repoURL,
		}, err.Error()))
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return errors.New(repoURL + T(" is not responding. Please make sure it is a valid plugin repo."))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(T("Error reading response from server: ") + err.Error())
	}

	result := clipr.PluginsJson{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return errors.New(T("Error processing data from server: ") + err.Error())
	}

	if result.Plugins == nil {
		return errors.New(T(`"Plugins" object not found in the responded data.`))
	}

	cmd.config.SetPluginRepo(models.PluginRepo{
		Name: c.Args()[0],
		URL:  c.Args()[1],
	})

	cmd.ui.Ok()
	cmd.ui.Say(repoURL + T(" added as '") + c.Args()[0] + "'")
	cmd.ui.Say("")
	return nil
}

func (cmd AddPluginRepo) checkIfRepoExists(repoName, repoURL string) error {
	repos := cmd.config.PluginRepos()
	for _, repo := range repos {
		if strings.ToLower(repo.Name) == strings.ToLower(repoName) {
			return errors.New(T(`Plugin repo named "{{.repoName}}" already exists, please use another name.`, map[string]interface{}{"repoName": repoName}))
		} else if repo.URL == repoURL {
			return errors.New(repo.URL + ` (` + repo.Name + T(`) already exists.`))
		}
	}
	return nil
}

func (cmd AddPluginRepo) verifyURL(repoURL string) (string, error) {
	if !strings.HasPrefix(repoURL, "http://") && !strings.HasPrefix(repoURL, "https://") {
		return "", errors.New(T("{{.URL}} is not a valid url, please provide a url, e.g. https://your_repo.com", map[string]interface{}{"URL": repoURL}))
	}

	if strings.HasSuffix(repoURL, "/") {
		repoURL = repoURL + "list"
	} else {
		repoURL = repoURL + "/list"
	}

	return repoURL, nil
}
