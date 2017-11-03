package pluginrepo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/cf/models"
	clipr "github.com/cloudfoundry/cli-plugin-repo/web"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

//go:generate counterfeiter . PluginRepo

type PluginRepo interface {
	GetPlugins([]models.PluginRepo) (map[string][]clipr.Plugin, []string)
}

type pluginRepo struct{}

func NewPluginRepo() PluginRepo {
	return pluginRepo{}
}

func (r pluginRepo) GetPlugins(repos []models.PluginRepo) (map[string][]clipr.Plugin, []string) {
	var pluginList clipr.PluginsJson
	repoError := []string{}
	repoPlugins := make(map[string][]clipr.Plugin)

	for _, repo := range repos {
		resp, err := http.Get(getListEndpoint(repo.URL))
		if err != nil {
			repoError = append(repoError, fmt.Sprintf(T("Error requesting from")+" '%s' - %s", repo.Name, err.Error()))
			continue
		} else {
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				repoError = append(repoError, fmt.Sprintf(T("Error reading response from")+" '%s' - %s ", repo.Name, err.Error()))
				continue
			}

			pluginList = clipr.PluginsJson{Plugins: nil}
			err = json.Unmarshal(body, &pluginList)
			if err != nil {
				repoError = append(repoError, fmt.Sprintf(T("Invalid json data from")+" '%s' - %s", repo.Name, err.Error()))
				continue
			} else if pluginList.Plugins == nil {
				repoError = append(repoError, T("Invalid data from '{{.repoName}}' - plugin data does not exist", map[string]interface{}{"repoName": repo.Name}))
				continue
			}

		}

		repoPlugins[repo.Name] = pluginList.Plugins
	}

	return repoPlugins, repoError
}

func getListEndpoint(url string) string {
	if strings.HasSuffix(url, "/") {
		return url + "list"
	}
	return url + "/list"
}
