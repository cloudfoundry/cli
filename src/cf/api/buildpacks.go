package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"net/url"
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
	return repo.findAllWithPath(path)
}

func (repo CloudControllerBuildpackRepository) FindByName(name string) (buildpack cf.Buildpack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s%s?q=name%%3A%s", repo.config.Target, buildpacks_path, url.QueryEscape(name))
	buildpacks, apiResponse := repo.findAllWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(buildpacks) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Buildpack", name)
		return
	}

	buildpack = buildpacks[0]
	return
}

func (repo CloudControllerBuildpackRepository) findAllWithPath(path string) (buildpacks []cf.Buildpack, apiResponse net.ApiResponse) {
	response := new(PaginatedBuildpackResources)

	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range response.Resources {
		buildpacks = append(buildpacks, unmarshallBuildpack(r))
	}

	return
}

func (repo CloudControllerBuildpackRepository) Create(newBuildpack cf.Buildpack) (createdBuildpack cf.Buildpack, apiResponse net.ApiResponse) {
	path := repo.config.Target + buildpacks_path
	entity := BuildpackEntity{Name: newBuildpack.Name, Priority: newBuildpack.Priority}
	body, err := json.Marshal(entity)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Could not serialize information", err)
		return
	}

	resource := new(BuildpackResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, bytes.NewReader(body), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdBuildpack = unmarshallBuildpack(*resource)
	return
}

func (repo CloudControllerBuildpackRepository) Delete(buildpack cf.Buildpack) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s%s/%s", repo.config.Target, buildpacks_path, buildpack.Guid)
	apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken)
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

	resource := new(BuildpackResource)
	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken, bytes.NewReader(body), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedBuildpack = unmarshallBuildpack(*resource)
	return
}

func unmarshallBuildpack(resource BuildpackResource) (buildpack cf.Buildpack) {
	buildpack.Guid = resource.Metadata.Guid
	buildpack.Name = resource.Entity.Name
	buildpack.Priority = resource.Entity.Priority
	return
}
