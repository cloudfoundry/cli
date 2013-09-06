package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"strings"
)

type SpaceRepository interface {
	GetCurrentSpace() (space cf.Space)
	FindAll() (spaces []cf.Space, err error)
	FindByName(name string) (space cf.Space, err error)
	GetSummary() (space cf.Space, err error)
}

type CloudControllerSpaceRepository struct {
	config *configuration.Configuration
}

func NewCloudControllerSpaceRepository(config *configuration.Configuration) (repo CloudControllerSpaceRepository) {
	repo.config = config
	return
}

func (repo CloudControllerSpaceRepository) GetCurrentSpace() (space cf.Space) {
	return repo.config.Space
}

func (repo CloudControllerSpaceRepository) FindAll() (spaces []cf.Space, err error) {
	path := fmt.Sprintf("%s/v2/organizations/%s/spaces", repo.config.Target, repo.config.Organization.Guid)
	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(ApiResponse)

	_, err = PerformRequestAndParseResponse(request, response)

	if err != nil {
		return
	}

	for _, r := range response.Resources {
		spaces = append(spaces, cf.Space{Name: r.Entity.Name, Guid: r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerSpaceRepository) FindByName(name string) (space cf.Space, err error) {
	spaces, err := repo.FindAll()
	lowerName := strings.ToLower(name)

	if err != nil {
		return
	}

	for _, s := range spaces {
		if strings.ToLower(s.Name) == lowerName {
			return s, nil
		}
	}

	err = errors.New("Space not found")
	return
}

func (repo CloudControllerSpaceRepository) GetSummary() (space cf.Space, err error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/summary", repo.config.Target, repo.config.Space.Guid)
	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(SpaceSummary) // but not an ApiResponse
	_, err = PerformRequestAndParseResponse(request, response)

	if err != nil {
		return
	}

	applications := extractApplicationsFromSummary(response.Apps)
	serviceInstances := extractServiceInstancesFromSummary(response.ServiceInstances, response.Apps)

	space = cf.Space{Name: response.Name, Guid: response.Guid, Applications: applications, ServiceInstances: serviceInstances}

	return
}

func extractApplicationsFromSummary(appSummaries []ApplicationSummary) (applications []cf.Application) {
	for _, appSummary := range appSummaries {
		app := cf.Application{
			Name:             appSummary.Name,
			Guid:             appSummary.Guid,
			Urls:             appSummary.Urls,
			State:            strings.ToLower(appSummary.State),
			Instances:        appSummary.Instances,
			RunningInstances: appSummary.RunningInstances,
			Memory:           appSummary.Memory,
		}
		applications = append(applications, app)
	}

	return
}

func extractServiceInstancesFromSummary(instanceSummaries []ServiceInstanceSummary, appSummaries []ApplicationSummary) (instances []cf.ServiceInstance) {
	for _, instanceSummary := range instanceSummaries {
		applicationNames := findApplicationNamesForInstance(instanceSummary.Name, appSummaries)

		planSummary := instanceSummary.ServicePlan
		offeringSummary := planSummary.ServiceOffering

		serviceOffering := cf.ServiceOffering{
			Label:    offeringSummary.Label,
			Provider: offeringSummary.Provider,
			Version:  offeringSummary.Version,
		}

		servicePlan := cf.ServicePlan{
			Name:            planSummary.Name,
			ServiceOffering: serviceOffering,
		}

		instance := cf.ServiceInstance{
			Name:             instanceSummary.Name,
			ServicePlan:      servicePlan,
			ApplicationNames: applicationNames,
		}

		instances = append(instances, instance)
	}

	return
}

func findApplicationNamesForInstance(instanceName string, appSummaries []ApplicationSummary) (applicationNames []string) {
	for _, appSummary := range appSummaries {
		for _, name := range appSummary.ServiceNames {
			if name == instanceName {
				applicationNames = append(applicationNames, appSummary.Name)
			}
		}
	}

	return
}
