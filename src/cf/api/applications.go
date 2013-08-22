package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type ApplicationRepository interface {
	FindByName(config *configuration.Configuration, name string) (app cf.Application, err error)
	SetEnv(config *configuration.Configuration, app cf.Application, name string, value string) (err error)
}

type CloudControllerApplicationRepository struct {
}

func (repo CloudControllerApplicationRepository) FindByName(config *configuration.Configuration, name string) (app cf.Application, err error) {
	apps, err := findApplications(config)
	lowerName := strings.ToLower(name)

	if err != nil {
		return
	}

	for _, a := range apps {
		if strings.ToLower(a.Name) == lowerName {
			return a, nil
		}
	}

	err = errors.New("Application not found")
	return
}

func (repo CloudControllerApplicationRepository) SetEnv(config *configuration.Configuration, app cf.Application, name string, value string) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s", config.Target, app.Guid)
	data := fmt.Sprintf(`{"environment_json":{"%s":"%s"}}`, name, value)
	request, err := http.NewRequest("PUT", path, strings.NewReader(data))
	request.Header.Set("Authorization", config.AccessToken)

	if err != nil {
		return
	}

	err = PerformRequest(request)
	return
}

func findApplications(config *configuration.Configuration) (apps []cf.Application, err error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps", config.Target, config.Space.Guid)
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return
	}
	request.Header.Set("Authorization", config.AccessToken)

	response := new(ApiResponse)

	err = PerformRequestForBody(request, response)

	if err != nil {
		return
	}

	for _, r := range response.Resources {
		apps = append(apps, cf.Application{r.Entity.Name, r.Metadata.Guid})
	}

	return
}
