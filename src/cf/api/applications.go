package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ApplicationRepository interface {
	FindByName(name string) (app cf.Application, apiStatus net.ApiStatus)
	SetEnv(app cf.Application, envVars map[string]string) (apiStatus net.ApiStatus)
	Create(newApp cf.Application) (createdApp cf.Application, apiStatus net.ApiStatus)
	Delete(app cf.Application) (apiStatus net.ApiStatus)
	Rename(app cf.Application, newName string) (apiStatus net.ApiStatus)
	Scale(app cf.Application) (apiStatus net.ApiStatus)
	Start(app cf.Application) (updatedApp cf.Application, apiStatus net.ApiStatus)
	Stop(app cf.Application) (updatedApp cf.Application, apiStatus net.ApiStatus)
	GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiStatus net.ApiStatus)
}

type CloudControllerApplicationRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerApplicationRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationRepository) FindByName(name string) (app cf.Application, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.Space.Guid, "%3A"+name)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	findResponse := new(ApplicationsApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, findResponse)
	if apiStatus.IsError() {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiStatus = net.NewNotFoundApiStatus()
		return
	}

	res := findResponse.Resources[0]
	path = fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, res.Metadata.Guid)
	request, apiStatus = repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	summaryResponse := new(ApplicationSummary)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, summaryResponse)
	if apiStatus.IsError() {
		return
	}

	urls := []string{}
	// This is a little wonky but we made a concious effort
	// to keep the domain very separate from the API repsonses
	// to maintain flexibility.
	domainRoute := cf.Route{}
	for _, route := range summaryResponse.Routes {
		domainRoute.Domain = cf.Domain{Name: route.Domain.Name}
		domainRoute.Host = route.Host
		urls = append(urls, domainRoute.URL())
	}

	app = cf.Application{
		Name:             summaryResponse.Name,
		Guid:             summaryResponse.Guid,
		Instances:        summaryResponse.Instances,
		RunningInstances: summaryResponse.RunningInstances,
		Memory:           summaryResponse.Memory,
		EnvironmentVars:  res.Entity.EnvironmentJson,
		Urls:             urls,
		State:            strings.ToLower(summaryResponse.State),
	}

	return
}

func (repo CloudControllerApplicationRepository) SetEnv(app cf.Application, envVars map[string]string) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	type setEnvReqBody struct {
		EnvJson map[string]string `json:"environment_json"`
	}

	body := setEnvReqBody{EnvJson: envVars}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Error creating json", err)
		return
	}

	request, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiStatus net.ApiStatus) {
	apiStatus = validateApplication(newApp)
	if apiStatus.IsError() {
		return
	}

	buildpackUrl := stringOrNull(newApp.BuildpackUrl)
	stackGuid := stringOrNull(newApp.Stack.Guid)
	command := stringOrNull(newApp.Command)

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	data := fmt.Sprintf(
		`{"space_guid":"%s","name":"%s","instances":%d,"buildpack":%s,"command":null,"memory":%d,"stack_guid":%s,"command":%s}`,
		repo.config.Space.Guid, newApp.Name, newApp.Instances, buildpackUrl, newApp.Memory, stackGuid, command,
	)
	request, apiStatus := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsError() {
		return
	}

	resource := new(Resource)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiStatus.IsError() {
		return
	}

	createdApp.Guid = resource.Metadata.Guid
	createdApp.Name = resource.Entity.Name
	return
}

func stringOrNull(s string) string {
	if s == "" {
		return "null"
	}

	return fmt.Sprintf(`"%s"`, s)
}

func (repo CloudControllerApplicationRepository) Delete(app cf.Application) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, app.Guid)
	request, apiStatus := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Rename(app cf.Application, newName string) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, newName)
	request, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Scale(app cf.Application) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	values := map[string]interface{}{}
	if app.DiskQuota > 0 {
		values["disk_quota"] = app.DiskQuota
	}
	if app.Instances > 0 {
		values["instances"] = app.Instances
	}
	if app.Memory > 0 {
		values["memory"] = app.Memory
	}

	bodyBytes, err := json.Marshal(values)
	if err != nil {
		return net.NewApiStatusWithError("Error generating body", err)
	}

	request, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(bodyBytes))
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiStatus net.ApiStatus) {
	updates := map[string]interface{}{"state": "STARTED"}
	if app.BuildpackUrl != "" {
		updates["buildpack"] = app.BuildpackUrl
	}
	return repo.updateApplication(app, updates)
}

func (repo CloudControllerApplicationRepository) Stop(app cf.Application) (updatedApp cf.Application, apiStatus net.ApiStatus) {
	return repo.updateApplication(app, map[string]interface{}{"state": "STOPPED"})
}

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
	Since float64
}

func (repo CloudControllerApplicationRepository) GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.Target, app.Guid)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	apiResponse := InstancesApiResponse{}

	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, &apiResponse)
	if apiStatus.IsError() {
		return
	}

	instances = make([]cf.ApplicationInstance, len(apiResponse), len(apiResponse))
	for k, v := range apiResponse {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}

		instances[index] = cf.ApplicationInstance{
			State: cf.InstanceState(strings.ToLower(v.State)),
			Since: time.Unix(int64(v.Since), 0),
		}
	}
	return
}

func (repo CloudControllerApplicationRepository) updateApplication(app cf.Application, updates map[string]interface{}) (updatedApp cf.Application, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	updates["console"] = true

	body, err := json.Marshal(updates)
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Could not serialize app updates.", err)
		return
	}

	request, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(body))

	if apiStatus.IsError() {
		return
	}

	response := ApplicationResource{}
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, &response)

	updatedApp = cf.Application{
		Name:  response.Entity.Name,
		Guid:  response.Metadata.Guid,
		State: strings.ToLower(response.Entity.State),
	}

	return
}

func validateApplication(app cf.Application) (apiStatus net.ApiStatus) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		apiStatus = net.NewApiStatusWithMessage("Application name is invalid. Name can only contain letters, numbers, underscores and hyphens.")
	}

	return
}
