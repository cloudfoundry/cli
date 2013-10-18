package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

const (
	buildpacks_path = "/v2/buildpacks"
)

type PaginatedBuildpackResources struct {
	Resources []BuildpackResource
}

type BuildpackResource struct {
	Resource
	Entity BuildpackEntity
}

type BuildpackEntity struct {
	Name     string `json:"name"`
	Priority *int   `json:"priority,omitempty"`
}

type BuildpackRepository interface {
	FindByName(name string) (buildpack cf.Buildpack, apiResponse net.ApiResponse)
	FindAll() (instances []cf.Buildpack, apiResponse net.ApiResponse)
	Create(newBuildpack cf.Buildpack) (createdBuildpack cf.Buildpack, apiResponse net.ApiResponse)
	Delete(buildpack cf.Buildpack) (apiResponse net.ApiResponse)
	Update(buildpack cf.Buildpack) (updatedBuildpack cf.Buildpack, apiResponse net.ApiResponse)
}

type CloudControllerBuildpackRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerBuildpackRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerBuildpackRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerBuildpackRepository) FindAll() (buildpacks []cf.Buildpack, apiResponse net.ApiResponse) {
	path := repo.config.Target + buildpacks_path
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}
	response := new(PaginatedBuildpackResources)

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		buildpacks = append(buildpacks, cf.Buildpack{
			Name:     r.Entity.Name,
			Guid:     r.Metadata.Guid,
			Priority: r.Entity.Priority,
		},
		)
	}

	return
}

func (repo CloudControllerBuildpackRepository) FindByName(name string) (buildpack cf.Buildpack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s%s?name=%s", repo.config.Target, buildpacks_path, url.QueryEscape(name))
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}
	buildpackResources := new(PaginatedBuildpackResources)

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, buildpackResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(buildpackResources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Buildpack", name)
		return
	}

	r := buildpackResources.Resources[0]

	buildpack = cf.Buildpack{
		Name:     r.Entity.Name,
		Guid:     r.Metadata.Guid,
		Priority: r.Entity.Priority,
	}

	return
}

func (repo CloudControllerBuildpackRepository) Create(newBuildpack cf.Buildpack) (createdBuildpack cf.Buildpack, apiResponse net.ApiResponse) {
	apiResponse = validateBuildpack(newBuildpack)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := repo.config.Target + buildpacks_path
	entity := BuildpackEntity{newBuildpack.Name, newBuildpack.Priority}
	body, err := json.Marshal(entity)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Could not serialize information", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, bytes.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	resource := new(BuildpackResource)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdBuildpack.Guid = resource.Metadata.Guid
	createdBuildpack.Name = resource.Entity.Name
	createdBuildpack.Priority = resource.Entity.Priority
	return
}

func (repo CloudControllerBuildpackRepository) Delete(buildpack cf.Buildpack) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s%s/%s", repo.config.Target, buildpacks_path, buildpack.Guid)
	request, apiResponse := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerBuildpackRepository) Update(buildpack cf.Buildpack) (updatedBuildpack cf.Buildpack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s%s/%s", repo.config.Target, buildpacks_path, buildpack.Guid)

	entity := BuildpackEntity{buildpack.Name, buildpack.Priority}
	body, err := json.Marshal(entity)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Could not serialize updates.", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	response := BuildpackResource{}
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, &response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedBuildpack = cf.Buildpack{
		Guid:     response.Metadata.Guid,
		Name:     response.Entity.Name,
		Priority: response.Entity.Priority,
	}

	return
}

func validateBuildpack(buildpack cf.Buildpack) (apiResponse net.ApiResponse) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(buildpack.Name) {
		apiResponse = net.NewApiResponseWithMessage("Buildpack name is invalid: name can only contain letters, numbers, underscores and hyphens")
	}

	return
}
