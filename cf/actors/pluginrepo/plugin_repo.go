package pluginrepo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	clipr "code.cloudfoundry.org/cli-plugin-repo/web"
	"code.cloudfoundry.org/cli/cf/models"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . PluginRepo

type PluginRepo interface {
	GetPlugins([]models.PluginRepo) (map[string][]clipr.Plugin, []string)
}

type pluginRepo struct{}

func NewPluginRepo() PluginRepo {
	return pluginRepo{}
}

func (r pluginRepo) GetPlugins(repos []models.PluginRepo) (map[string][]clipr.Plugin, []string) {
	var pluginList clipr.PluginsJson
	var repoError []string
	repoPlugins := make(map[string][]clipr.Plugin)

	for _, repo := range repos {
		// resp, err := http.Get(getListEndpoint(repo.URL))
		client := &http.Client{}
		req, err := http.NewRequest("GET", getListEndpoint(repo.URL), nil)
		if err != nil {
			repoError = append(repoError, fmt.Sprintf(T("Error requesting from")+" '%s' - %s", repo.Name, err.Error()))
			continue
		}

		req.Header.Set("User-Agent", "golang_user_agent/1.0")
		resp, err := client.Do(req)
		if err != nil {
			repoError = append(repoError, fmt.Sprintf(T("Error requesting from")+" '%s' - %s", repo.Name, err.Error()))
			continue
		} else {
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					repoError = append(repoError, fmt.Sprintf(T("Error closing body")+" '%s' - %s", repo.Name, err.Error()))
				}
			}(resp.Body)

			body, err := io.ReadAll(resp.Body)
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
