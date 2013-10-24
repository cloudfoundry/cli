package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PaginatedApplicationResources struct {
	Resources []ApplicationResource
}

type ApplicationResource struct {
	Resource
	Entity ApplicationEntity
}

type ApplicationEntity struct {
	Name            string
	State           string
	Instances       int
	Memory          int
	Routes          []AppRouteResource
	EnvironmentJson map[string]string `json:"environment_json"`
}

type AppRouteResource struct {
	Resource
	Entity AppRouteEntity
}

type AppRouteEntity struct {
	Host   string
	Domain Resource
}

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
	appResources := new(PaginatedApplicationResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, appResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(appResources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "App", name)
		return
	}

	res := appResources.Resources[0]
	app = cf.Application{
		Guid:            res.Metadata.Guid,
		Name:            res.Entity.Name,
		EnvironmentVars: res.Entity.EnvironmentJson,
		State:           res.Entity.State,
		Instances:       res.Entity.Instances,
		Memory:          uint64(res.Entity.Memory),
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

	apiResponse = repo.gateway.UpdateResource(path, repo.config.AccessToken, bytes.NewReader(jsonBytes))
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

	resource := new(Resource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdApp.Guid = resource.Metadata.Guid
	createdApp.Name = resource.Entity.Name
	return
}

func (repo CloudControllerApplicationRepository) Delete(app cf.Application) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, app.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerApplicationRepository) Rename(app cf.Application, newName string) (apiResponse net.ApiResponse) {
	app.Name = newName
	data := fmt.Sprintf(`{"name":"%s"}`, newName)
	apiResponse = repo.updateApp(app, strings.NewReader(data))
	return
}

func (repo CloudControllerApplicationRepository) Scale(app cf.Application) (apiResponse net.ApiResponse) {
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

	apiResponse = repo.updateApp(app, bytes.NewReader(bodyBytes))
	return
}

func (repo CloudControllerApplicationRepository) updateApp(app cf.Application, body io.ReadSeeker) (apiResponse net.ApiResponse) {
	apiResponse = validateApplication(app)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, body)
}

func validateApplication(app cf.Application) (apiResponse net.ApiResponse) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		apiResponse = net.NewApiResponseWithMessage("App name is invalid: name can only contain letters, numbers, underscores and hyphens")
	}

	return
}

func (repo CloudControllerApplicationRepository) Start(app cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	updates := map[string]interface{}{"state": "STARTED"}
	if app.BuildpackUrl != "" {
		updates["buildpack"] = app.BuildpackUrl
	}
	return repo.startOrStopApp(app, updates)
}

func (repo CloudControllerApplicationRepository) Stop(app cf.Application) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	return repo.startOrStopApp(app, map[string]interface{}{"state": "STOPPED"})
}

func (repo CloudControllerApplicationRepository) startOrStopApp(app cf.Application, updates map[string]interface{}) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)

	updates["console"] = true

	body, err := json.Marshal(updates)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Could not serialize app updates.", err)
		return
	}

	resource := new(ApplicationResource)
	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken, bytes.NewReader(body), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedApp = cf.Application{
		Name:  resource.Entity.Name,
		Guid:  resource.Metadata.Guid,
		State: strings.ToLower(resource.Entity.State),
	}

	return
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
