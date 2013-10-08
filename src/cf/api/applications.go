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
	FindByName(name string) (app cf.Application, apiResponse net.ApiResponse)
	SetEnv(app cf.Application, envVars map[string]string) (apiResponse net.ApiResponse)
	Create(newApp cf.Application) (createdApp cf.Application, apiResponse net.ApiResponse)
	Delete(app cf.Application) (apiResponse net.ApiResponse)
	Rename(app cf.Application, newName string) (apiResponse net.ApiResponse)
	Scale(app cf.Application) (apiResponse net.ApiResponse)
	Start(app cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse)
	Stop(app cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse)
	GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiResponse net.ApiResponse)
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

func (repo CloudControllerApplicationRepository) FindByName(name string) (app cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.Space.Guid, "%3A"+name)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	findResponse := new(ApplicationsApiResponse)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, findResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("App", name)
		return
	}

	res := findResponse.Resources[0]
	path = fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, res.Metadata.Guid)
	request, apiResponse = repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	summaryResponse := new(ApplicationSummary)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, summaryResponse)
	if apiResponse.IsNotSuccessful() {
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

func (repo CloudControllerApplicationRepository) SetEnv(app cf.Application, envVars map[string]string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	type setEnvReqBody struct {
		EnvJson map[string]string `json:"environment_json"`
	}

	body := setEnvReqBody{EnvJson: envVars}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error creating json", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiResponse net.ApiResponse) {
	apiResponse = validateApplication(newApp)
	if apiResponse.IsNotSuccessful() {
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
	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	resource := new(Resource)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiResponse.IsNotSuccessful() {
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

func (repo CloudControllerApplicationRepository) Delete(app cf.Application) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, app.Guid)
	request, apiResponse := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Rename(app cf.Application, newName string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	data := fmt.Sprintf(`{"name":"%s"}`, newName)
	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Scale(app cf.Application) (apiResponse net.ApiResponse) {
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
		return net.NewApiResponseWithError("Error generating body", err)
	}

	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(bodyBytes))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	updates := map[string]interface{}{"state": "STARTED"}
	if app.BuildpackUrl != "" {
		updates["buildpack"] = app.BuildpackUrl
	}
	return repo.updateApplication(app, updates)
}

func (repo CloudControllerApplicationRepository) Stop(app cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	return repo.updateApplication(app, map[string]interface{}{"state": "STOPPED"})
}

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
	Since float64
}

func (repo CloudControllerApplicationRepository) GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.Target, app.Guid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instancesResponse := InstancesApiResponse{}

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &instancesResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	instances = make([]cf.ApplicationInstance, len(instancesResponse), len(instancesResponse))
	for k, v := range instancesResponse {
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

func (repo CloudControllerApplicationRepository) updateApplication(app cf.Application, updates map[string]interface{}) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	updates["console"] = true

	body, err := json.Marshal(updates)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Could not serialize app updates.", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(body))

	if apiResponse.IsNotSuccessful() {
		return
	}

	response := ApplicationResource{}
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &response)

	updatedApp = cf.Application{
		Name:  response.Entity.Name,
		Guid:  response.Metadata.Guid,
		State: strings.ToLower(response.Entity.State),
	}

	return
}

func validateApplication(app cf.Application) (apiResponse net.ApiResponse) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		apiResponse = net.NewApiResponseWithMessage("App name is invalid: name can only contain letters, numbers, underscores and hyphens")
	}

	return
}
