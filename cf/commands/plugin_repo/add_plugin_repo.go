package plugin_repo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type AddPluginRepo struct {
	ui     terminal.UI
	config core_config.ReadWriter
}

func NewAddPluginRepo(ui terminal.UI, config core_config.ReadWriter) AddPluginRepo {
	return AddPluginRepo{
		ui:     ui,
		config: config,
	}
}

func (cmd AddPluginRepo) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "add-plugin-repo",
		Description: T("Add a new plugin repository"),
		Usage: T(`CF_NAME add-plugin-repo [REPO_NAME] [URL]

EXAMPLE:
   cf add-plugin-repo PrivateRepo http://myprivaterepo.com/repo/
`),
		TotalArgs: 2,
	}
}

func (cmd AddPluginRepo) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}
	return
}

func (cmd AddPluginRepo) Run(c *cli.Context) {

	cmd.ui.Say("")
	repoUrl := strings.ToLower(c.Args()[1])
	repoName := strings.Trim(c.Args()[0], " ")

	cmd.checkIfRepoExists(repoName, repoUrl)

	repoUrl = cmd.verifyUrl(repoUrl)

	resp, err := http.Get(repoUrl)
	if err != nil {
		cmd.ui.Failed(T("There is an error performing request on '{{.repoUrl}}': ", map[string]interface{}{"repoUrl": repoUrl}), err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		cmd.ui.Failed(repoUrl + T(" is not responding. Please make sure it is a valid plugin repo."))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		cmd.ui.Failed(T("Error reading response from server: ") + err.Error())
	}

	result := clipr.PluginsJson{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		cmd.ui.Failed(T("Error processing data from server: ") + err.Error())
	}

	if result.Plugins == nil {
		cmd.ui.Failed(T(`"Plugins" object not found in the responded data.`))
	}

	cmd.config.SetPluginRepo(models.PluginRepo{
		Name: c.Args()[0],
		Url:  c.Args()[1],
	})

	cmd.ui.Ok()
	cmd.ui.Say(repoUrl + T(" added as '") + c.Args()[0] + "'")
	cmd.ui.Say("")
}

func (cmd AddPluginRepo) checkIfRepoExists(repoName, repoUrl string) {
	repos := cmd.config.PluginRepos()
	for _, repo := range repos {
		if strings.ToLower(repo.Name) == strings.ToLower(repoName) {
			cmd.ui.Failed(T(`Plugin repo named "{{.repoName}}" already exists, please use another name.`, map[string]interface{}{"repoName": repoName}))
		} else if repo.Url == repoUrl {
			cmd.ui.Failed(repo.Url + ` (` + repo.Name + T(`) already exists.`))
		}
	}
}

func (cmd AddPluginRepo) verifyUrl(repoUrl string) string {
	if !strings.HasPrefix(repoUrl, "http://") && !strings.HasPrefix(repoUrl, "https://") {
		cmd.ui.Failed(repoUrl + T(" is not a valid url, please provide a url, e.g. http://your_repo.com"))
	}

	if strings.HasSuffix(repoUrl, "/") {
		repoUrl = repoUrl + "list"
	} else {
		repoUrl = repoUrl + "/list"
	}

	return repoUrl
}
