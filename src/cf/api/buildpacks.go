package api

import (
	"bytes"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"encoding/json"
	"fmt"
	"net/url"
)

type BuildpackRepository interface {
	FindByName(name string) (buildpack models.Buildpack, apiResponse net.ApiResponse)
	ListBuildpacks(func([]models.Buildpack) bool) net.ApiResponse
	Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiResponse net.ApiResponse)
	Delete(buildpackGuid string) (apiResponse net.ApiResponse)
	Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiResponse net.ApiResponse)
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

func (repo CloudControllerBuildpackRepository) ListBuildpacks(cb func([]models.Buildpack) bool) net.ApiResponse {
	return repo.gateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		buildpacks_path,
		&BuildpackResource{},
		func(results []interface{}) bool {
			buildpacks := make([]models.Buildpack, 0, len(results))
			for _, result := range results {
				buildpacks = append(buildpacks, result.(models.Buildpack))
			}
			return cb(buildpacks)
		})
}

func (repo CloudControllerBuildpackRepository) FindByName(name string) (buildpack models.Buildpack, apiResponse net.ApiResponse) {
	foundIt := false
	apiResponse = repo.gateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		fmt.Sprintf("%s?q=name%%3A%s", buildpacks_path, url.QueryEscape(name)),
		&BuildpackResource{},
		func(results []interface{}) bool {
			buildpack = results[0].(models.Buildpack)
			foundIt = true
			return false
		})

	if !foundIt {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Buildpack", name)
	}
	return
}

func (repo CloudControllerBuildpackRepository) Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiResponse net.ApiResponse) {
	path := repo.config.Target + buildpacks_path
	entity := BuildpackEntity{Name: name, Position: position, Enabled: enabled, Locked: locked}
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

	createdBuildpack = resource.ToFields().(models.Buildpack)
	return
}

func (repo CloudControllerBuildpackRepository) Delete(buildpackGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s%s/%s", repo.config.Target, buildpacks_path, buildpackGuid)
	apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken)
	return
}

func (repo CloudControllerBuildpackRepository) Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s%s/%s", repo.config.Target, buildpacks_path, buildpack.Guid)

	entity := BuildpackEntity{buildpack.Name, buildpack.Position, buildpack.Enabled, "", "", buildpack.Locked}

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

	updatedBuildpack = resource.ToFields().(models.Buildpack)
	return
}

const buildpacks_path = "/v2/buildpacks"

func (resource *BuildpackResource) ToFields() interface{} {
	return models.Buildpack{
		Guid:     resource.Metadata.Guid,
		Name:     resource.Entity.Name,
		Position: resource.Entity.Position,
		Enabled:  resource.Entity.Enabled,
		Key:      resource.Entity.Key,
		Filename: resource.Entity.Filename,
		Locked:   resource.Entity.Locked,
	}
}

type BuildpackResource struct {
	Metadata Metadata
	Entity   BuildpackEntity
}

type BuildpackEntity struct {
	Name     string `json:"name"`
	Position *int   `json:"position,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Key      string `json:"key,omitempty"`
	Filename string `json:"filename,omitempty"`
	Locked   *bool  `json:"locked,omitempty"`
}
