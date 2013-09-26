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
	FindByName(name string) (app cf.Application, found bool, apiErr *net.ApiError)
	SetEnv(app cf.Application, envVars map[string]string) (apiErr *net.ApiError)
	Create(newApp cf.Application) (createdApp cf.Application, apiErr *net.ApiError)
	Delete(app cf.Application) (apiErr *net.ApiError)
	Rename(app cf.Application, newName string) (apiErr *net.ApiError)
	Scale(app cf.Application) (apiErr *net.ApiError)
	Start(app cf.Application) (updatedApp cf.Application, apiErr *net.ApiError)
	Stop(app cf.Application) (updatedApp cf.Application, apiErr *net.ApiError)
	GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiErr *net.ApiError)
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

func (repo CloudControllerApplicationRepository) FindByName(name string) (app cf.Application, found bool, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.Space.Guid, "%3A"+name)
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	findResponse := new(ApplicationsApiResponse)
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, findResponse)
	if apiErr != nil {
		return
	}

	if len(findResponse.Resources) == 0 {
		return
	}

	res := findResponse.Resources[0]
	path = fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, res.Metadata.Guid)
	request, apiErr = repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	summaryResponse := new(ApplicationSummary)
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, summaryResponse)
	if apiErr != nil {
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

	found = true

	return
}

func (repo CloudControllerApplicationRepository) SetEnv(app cf.Application, envVars map[string]string) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	type setEnvReqBody struct {
		EnvJson map[string]string `json:"environment_json"`
	}

	body := setEnvReqBody{EnvJson: envVars}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		apiErr = net.NewApiErrorWithError("Error creating json", err)
		return
	}

	request, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiErr *net.ApiError) {
	apiErr = validateApplication(newApp)
	if apiErr != nil {
		return
	}

	buildpackUrl := stringOrNull(newApp.BuildpackUrl)
	stackGuid := stringOrNull(newApp.Stack.Guid)

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	data := fmt.Sprintf(
		`{"space_guid":"%s","name":"%s","instances":%d,"buildpack":%s,"command":null,"memory":%d,"stack_guid":%s}`,
		repo.config.Space.Guid, newApp.Name, newApp.Instances, buildpackUrl, newApp.Memory, stackGuid,
	)
	request, apiErr := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	resource := new(Resource)
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiErr != nil {
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

func (repo CloudControllerApplicationRepository) Delete(app cf.Application) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, app.Guid)
	request, apiErr := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Rename(app cf.Application, newName string) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, newName)
	request, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Scale(app cf.Application) (apiErr *net.ApiError) {
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
		return net.NewApiErrorWithError("Error generating body", err)
	}

	request, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(bodyBytes))
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiErr *net.ApiError) {
	return repo.changeApplicationState(app, "STARTED")
}

func (repo CloudControllerApplicationRepository) Stop(app cf.Application) (updatedApp cf.Application, apiErr *net.ApiError) {
	return repo.changeApplicationState(app, "STOPPED")
}

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
	Since float64
}

func (repo CloudControllerApplicationRepository) GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.Target, app.Guid)
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiResponse := InstancesApiResponse{}

	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, &apiResponse)
	if apiErr != nil {
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

func (repo CloudControllerApplicationRepository) changeApplicationState(app cf.Application, state string) (updatedApp cf.Application, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	body := fmt.Sprintf(`{"console":true,"state":"%s"}`, state)
	request, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))

	if apiErr != nil {
		return
	}

	response := ApplicationResource{}
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, &response)

	updatedApp = cf.Application{
		Name:  response.Entity.Name,
		Guid:  response.Metadata.Guid,
		State: strings.ToLower(response.Entity.State),
	}

	return
}

func validateApplication(app cf.Application) (apiErr *net.ApiError) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		apiErr = net.NewApiErrorWithMessage("Application name is invalid. Name can only contain letters, numbers, underscores and hyphens.")
	}

	return
}
