package api

import (
    "cf"
    "cf/configuration"
    "cf/net"
    "fmt"
    "net/url"
)

type BuildpackRepository interface {
    FindByName(name string) (buildpack cf.Buildpack, apiResponse net.ApiResponse)
    FindAll() (instances []cf.Buildpack, apiResponse net.ApiResponse)
    // Create(buildpackToCreate cf.Buildpack) (createdBuildpack cf.Buildpack, apiResponse net.ApiResponse)
    // Delete(buildpack cf.Buildpack) (apiResponse net.ApiResponse)
    // Update(buildpack cf.Buildpack) (apiResponse net.ApiResponse)
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
    path := repo.config.Target + "/v2/buildpacks"
    request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
    if apiResponse.IsNotSuccessful() {
        return
    }
    response := new(PaginatedOrganizationResources)

    _, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, response)
    if apiResponse.IsNotSuccessful() {
        return
    }

    for _, r := range response.Resources {
        buildpacks = append(buildpacks, cf.Buildpack{
            Name: r.Entity.Name,
            Guid: r.Metadata.Guid,
        },
        )
    }

    return
}

func (repo CloudControllerBuildpackRepository) FindByName(name string) (buildpack cf.Buildpack, apiResponse net.ApiResponse) {
    path := fmt.Sprintf("%s/v2/buildpacks?name=%s", repo.config.Target, url.QueryEscape(name))
    request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
    if apiResponse.IsNotSuccessful() {
        return
    }
    buildpackResources := new(PaginatedOrganizationResources)

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
        Name: r.Entity.Name,
        Guid: r.Metadata.Guid,
    }

    return
}
