package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/models"
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
	NextUrl   string `json:"next_url"`
}

type BuildpackResource struct {
	Resource
	Entity BuildpackEntity
}

type BuildpackEntity struct {
	Name     string `json:"name"`
	Position *int   `json:"position,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Key      string `json:"key,omitempty"`
	Filename string `json:"filename,omitempty"`
	Locked   *bool  `json:"locked,omitempty"`
}

type BuildpackRepository interface {
	FindByName(name string) (buildpack models.Buildpack, apiResponse net.ApiResponse)
	ListBuildpacks(stop chan bool) (buildpacksChan chan []models.Buildpack, statusChan chan net.ApiResponse)
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

func (repo CloudControllerBuildpackRepository) ListBuildpacks(stop chan bool) (buildpacksChan chan []models.Buildpack, statusChan chan net.ApiResponse) {
	buildpacksChan = make(chan []models.Buildpack, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		path := buildpacks_path

	loop:
		for path != "" {
			select {
			case <-stop:
				break loop
			default:
				var (
					buildpacks  []models.Buildpack
					apiResponse net.ApiResponse
				)
				buildpacks, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(buildpacksChan)
					close(statusChan)
					return
				}

				if len(buildpacks) > 0 {
					buildpacksChan <- buildpacks
				}
			}
		}
		close(buildpacksChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerBuildpackRepository) FindByName(name string) (buildpack models.Buildpack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s?q=name%%3A%s", buildpacks_path, url.QueryEscape(name))
	buildpacks, _, apiResponse := repo.findNextWithPath(path)
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

func (repo CloudControllerBuildpackRepository) findNextWithPath(path string) (buildpacks []models.Buildpack, nextUrl string, apiResponse net.ApiResponse) {
	response := new(PaginatedBuildpackResources)

	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, response)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = response.NextUrl

	for _, r := range response.Resources {
		buildpacks = append(buildpacks, unmarshallBuildpack(r))
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

	createdBuildpack = unmarshallBuildpack(*resource)
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

	updatedBuildpack = unmarshallBuildpack(*resource)
	return
}

func unmarshallBuildpack(resource BuildpackResource) (buildpack models.Buildpack) {
	buildpack.Guid = resource.Metadata.Guid
	buildpack.Name = resource.Entity.Name
	buildpack.Position = resource.Entity.Position
	buildpack.Enabled = resource.Entity.Enabled
	buildpack.Key = resource.Entity.Key
	buildpack.Filename = resource.Entity.Filename
	buildpack.Locked = resource.Entity.Locked
	return
}
