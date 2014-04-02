package api

import (
	"bytes"
	"cf/api/resources"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"encoding/json"
	"fmt"
	"net/url"
)

type BuildpackRepository interface {
	FindByName(name string) (buildpack models.Buildpack, apiErr error)
	ListBuildpacks(func(models.Buildpack) bool) error
	Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiErr error)
	Delete(buildpackGuid string) (apiErr error)
	Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiErr error)
	Rename(buildpack models.Buildpack, newbuildpackName string) (renamedBuildpack models.Buildpack, apiErr error)
}

type CloudControllerBuildpackRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerBuildpackRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerBuildpackRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerBuildpackRepository) ListBuildpacks(cb func(models.Buildpack) bool) error {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		buildpacks_path,
		resources.BuildpackResource{},
		func(resource interface{}) bool {
			return cb(resource.(resources.BuildpackResource).ToFields())
		})
}

func (repo CloudControllerBuildpackRepository) FindByName(name string) (buildpack models.Buildpack, apiErr error) {
	foundIt := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("%s?q=%s", buildpacks_path, url.QueryEscape("name:"+name)),
		resources.BuildpackResource{},
		func(resource interface{}) bool {
			buildpack = resource.(resources.BuildpackResource).ToFields()
			foundIt = true
			return false
		})

	if !foundIt {
		apiErr = errors.NewModelNotFoundError("Buildpack", name)
	}
	return
}

func (repo CloudControllerBuildpackRepository) Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiErr error) {
	path := repo.config.ApiEndpoint() + buildpacks_path
	entity := resources.BuildpackEntity{Name: name, Position: position, Enabled: enabled, Locked: locked}
	body, err := json.Marshal(entity)
	if err != nil {
		apiErr = errors.NewWithError("Could not serialize information", err)
		return
	}

	resource := new(resources.BuildpackResource)
	apiErr = repo.gateway.CreateResource(path, bytes.NewReader(body), resource)
	if apiErr != nil {
		return
	}

	createdBuildpack = resource.ToFields()
	return
}

func (repo CloudControllerBuildpackRepository) Delete(buildpackGuid string) (apiErr error) {
	path := fmt.Sprintf("%s%s/%s", repo.config.ApiEndpoint(), buildpacks_path, buildpackGuid)
	apiErr = repo.gateway.DeleteResource(path)
	return
}

func (repo CloudControllerBuildpackRepository) Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiErr error) {
	path := fmt.Sprintf("%s%s/%s", repo.config.ApiEndpoint(), buildpacks_path, buildpack.Guid)

	entity := resources.BuildpackEntity{
		Name:     buildpack.Name,
		Position: buildpack.Position,
		Enabled:  buildpack.Enabled,
		Key:      "",
		Filename: "",
		Locked:   buildpack.Locked,
	}

	body, err := json.Marshal(entity)
	if err != nil {
		apiErr = errors.NewWithError("Could not serialize updates.", err)
		return
	}

	resource := new(resources.BuildpackResource)
	apiErr = repo.gateway.UpdateResource(path, bytes.NewReader(body), resource)
	if apiErr != nil {
		return
	}

	updatedBuildpack = resource.ToFields()
	return
}

func (repo CloudControllerBuildpackRepository) Rename(buildpack models.Buildpack, newbuildpackName string) (updatedBuildpack models.Buildpack, apiErr error) {
	buildpack.Name = newbuildpackName
	updatedBuildpack, apiErr = repo.Update(buildpack)
	return
}

const buildpacks_path = "/v2/buildpacks"
