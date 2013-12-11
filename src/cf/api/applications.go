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
	"strings"
)

type PaginatedApplicationResources struct {
	Resources []ApplicationResource
}

type ApplicationResource struct {
	Resource
	Entity ApplicationEntity
}

func (resource ApplicationResource) ToFields() (app cf.ApplicationFields) {
	app.Guid = resource.Metadata.Guid
	app.Name = resource.Entity.Name
	app.EnvironmentVars = resource.Entity.EnvironmentJson
	app.State = strings.ToLower(resource.Entity.State)
	app.InstanceCount = resource.Entity.Instances
	app.Memory = uint64(resource.Entity.Memory)

	return
}

func (resource ApplicationResource) ToModel() (app cf.Application) {
	app.ApplicationFields = resource.ToFields()

	for _, routeResource := range resource.Entity.Routes {
		app.Routes = append(app.Routes, routeResource.ToModel())
	}
	return
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

func (resource AppRouteResource) ToFields() (route cf.RouteFields) {
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	return
}

func (resource AppRouteResource) ToModel() (route cf.RouteSummary) {
	route.RouteFields = resource.ToFields()
	route.Domain.Guid = resource.Entity.Domain.Metadata.Guid
	route.Domain.Name = resource.Entity.Domain.Entity.Name
	return
}

type AppRouteEntity struct {
	Host   string
	Domain Resource
}

type ApplicationRepository interface {
	FindByName(name string) (app cf.Application, apiResponse net.ApiResponse)
	SetEnv(appGuid string, envVars map[string]string) (apiResponse net.ApiResponse)
	Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse)
	Update(app cf.ApplicationFields, stackGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse)
	Delete(appGuid string) (apiResponse net.ApiResponse)
	Rename(appGuid string, newName string) (apiResponse net.ApiResponse)
	Scale(app cf.ApplicationFields) (apiResponse net.ApiResponse)
	Start(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse)
	Stop(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse)
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
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.SpaceFields.Guid, "%3A"+name)
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
	app = res.ToModel()
	return
}
func (repo CloudControllerApplicationRepository) SetEnv(appGuid string, envVars map[string]string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, appGuid)

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

func (repo CloudControllerApplicationRepository) Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse) {
	data, apiResponse := repo.formatAppJSON(name, buildpackUrl, stackGuid, repo.config.SpaceFields.Guid, command, memory, instances)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdApp = resource.ToModel()
	return
}
func (repo CloudControllerApplicationRepository) formatAppJSON(name, buildpackUrl, stackGuid, spaceGuid, command string, memory uint64, instances int) (data string, apiResponse net.ApiResponse) {
	apiResponse = validateApplicationName(name)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if command == "null" {
		command = "\"\""
	} else {
		command = stringOrNull(command)
	}

	data = "{"
	if spaceGuid != "" {
		data += fmt.Sprintf(`"space_guid":"%s",`, spaceGuid)
	}
	data += fmt.Sprintf(
		`"name":"%s","instances":%d,"buildpack":%s,"memory":%d,"stack_guid":%s,"command":%s`,
		name,
		instances,
		stringOrNull(buildpackUrl),
		memory,
		stringOrNull(stackGuid),
		command,
	)
	data += "}"
	return
}

func (repo CloudControllerApplicationRepository) Update(app cf.ApplicationFields, stackGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	data, apiResponse := repo.formatAppJSON(app.Name, app.BuildpackUrl, stackGuid, "", app.Command, app.Memory, app.InstanceCount)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, appGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerApplicationRepository) Rename(appGuid, newName string) (apiResponse net.ApiResponse) {
	apiResponse = validateApplicationName(newName)
	if apiResponse.IsNotSuccessful() {
		return
	}

	data := fmt.Sprintf(`{"name":"%s"}`, newName)
	apiResponse = repo.updateApp(appGuid, strings.NewReader(data))
	return
}

func (repo CloudControllerApplicationRepository) Scale(app cf.ApplicationFields) (apiResponse net.ApiResponse) {
	values := map[string]interface{}{}
	if app.DiskQuota > 0 {
		values["disk_quota"] = app.DiskQuota
	}
	if app.InstanceCount > 0 {
		values["instances"] = app.InstanceCount
	}
	if app.Memory > 0 {
		values["memory"] = app.Memory
	}

	bodyBytes, err := json.Marshal(values)
	if err != nil {
		return net.NewApiResponseWithError("Error generating body", err)
	}

	apiResponse = repo.updateApp(app.Guid, bytes.NewReader(bodyBytes))
	return
}

func (repo CloudControllerApplicationRepository) updateApp(appGuid string, body io.ReadSeeker) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, appGuid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, body)
}

func validateApplicationName(name string) (apiResponse net.ApiResponse) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(name) {
		apiResponse = net.NewApiResponseWithMessage("App name is invalid: name can only contain letters, numbers, underscores and hyphens")
	}

	return
}

func (repo CloudControllerApplicationRepository) Start(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	return repo.startOrStopApp(appGuid, map[string]interface{}{"state": "STARTED"})
}

func (repo CloudControllerApplicationRepository) StartWithDifferentBuildpack(appGuid, buildpack string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	updates := map[string]interface{}{
		"state":     "STARTED",
		"buildpack": buildpack,
	}
	return repo.startOrStopApp(appGuid, updates)
}

func (repo CloudControllerApplicationRepository) Stop(appGuid string) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	return repo.startOrStopApp(appGuid, map[string]interface{}{"state": "STOPPED"})
}

func (repo CloudControllerApplicationRepository) startOrStopApp(appGuid string, updates map[string]interface{}) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?inline-relations-depth=2", repo.config.Target, appGuid)

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

	updatedApp = resource.ToModel()
	return
}
