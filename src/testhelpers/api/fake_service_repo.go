package api

import (
	"cf/models"
	"cf/net"
	"generic"
)

type FakeServiceRepo struct {
	GetAllServiceOfferingsReturns struct {
		ServiceOfferings []models.ServiceOffering
		ApiResponse      net.ApiResponse
	}

	GetServiceOfferingsForSpaceReturns struct {
		ServiceOfferings []models.ServiceOffering
		ApiResponse      net.ApiResponse
	}
	GetServiceOfferingsForSpaceArgs struct {
		SpaceGuid string
	}

	CreateServiceInstanceName     string
	CreateServiceInstancePlanGuid string
	CreateServiceAlreadyExists    bool

	FindInstanceByNameName            string
	FindInstanceByNameServiceInstance models.ServiceInstance
	FindInstanceByNameErr             bool
	FindInstanceByNameNotFound        bool

	FindInstanceByNameMap generic.Map

	DeleteServiceServiceInstance models.ServiceInstance

	RenameServiceServiceInstance models.ServiceInstance
	RenameServiceNewName         string

	PurgedServiceOffering           models.ServiceOffering
	PurgeServiceOfferingCalled      bool
	PurgeServiceOfferingApiResponse net.ApiResponse

	FindServiceOfferingByLabelAndProviderName            string
	FindServiceOfferingByLabelAndProviderProvider        string
	FindServiceOfferingByLabelAndProviderServiceOffering models.ServiceOffering
	FindServiceOfferingByLabelAndProviderApiResponse     net.ApiResponse
	FindServiceOfferingByLabelAndProviderCalled          bool
}

func (repo *FakeServiceRepo) GetAllServiceOfferings() (models.ServiceOfferings, net.ApiResponse) {
	return repo.GetAllServiceOfferingsReturns.ServiceOfferings, repo.GetAllServiceOfferingsReturns.ApiResponse
}

func (repo *FakeServiceRepo) GetServiceOfferingsForSpace(spaceGuid string) (models.ServiceOfferings, net.ApiResponse) {
	repo.GetServiceOfferingsForSpaceArgs.SpaceGuid = spaceGuid
	return repo.GetServiceOfferingsForSpaceReturns.ServiceOfferings, repo.GetServiceOfferingsForSpaceReturns.ApiResponse
}

func (repo *FakeServiceRepo) PurgeServiceOffering(offering models.ServiceOffering) (apiResponse net.ApiResponse) {
	repo.PurgedServiceOffering = offering
	repo.PurgeServiceOfferingCalled = true
	return
}

func (repo *FakeServiceRepo) FindServiceOfferingByLabelAndProvider(name, provider string) (offering models.ServiceOffering, apiResponse net.ApiResponse) {
	repo.FindServiceOfferingByLabelAndProviderCalled = true
	repo.FindServiceOfferingByLabelAndProviderName = name
	repo.FindServiceOfferingByLabelAndProviderProvider = provider
	apiResponse = repo.FindServiceOfferingByLabelAndProviderApiResponse
	offering = repo.FindServiceOfferingByLabelAndProviderServiceOffering
	return
}

func (repo *FakeServiceRepo) CreateServiceInstance(name, planGuid string) (identicalAlreadyExists bool, apiResponse net.ApiResponse) {
	repo.CreateServiceInstanceName = name
	repo.CreateServiceInstancePlanGuid = planGuid
	identicalAlreadyExists = repo.CreateServiceAlreadyExists

	return
}

func (repo *FakeServiceRepo) FindInstanceByName(name string) (instance models.ServiceInstance, apiResponse net.ApiResponse) {
	repo.FindInstanceByNameName = name

	if repo.FindInstanceByNameMap != nil && repo.FindInstanceByNameMap.Has(name) {
		instance = repo.FindInstanceByNameMap.Get(name).(models.ServiceInstance)
	} else {
		instance = repo.FindInstanceByNameServiceInstance
	}

	if repo.FindInstanceByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding instance")
	}

	if repo.FindInstanceByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Service instance", name)
	}

	return
}

func (repo *FakeServiceRepo) DeleteService(instance models.ServiceInstance) (apiResponse net.ApiResponse) {
	repo.DeleteServiceServiceInstance = instance
	return
}

func (repo *FakeServiceRepo) RenameService(instance models.ServiceInstance, newName string) (apiResponse net.ApiResponse) {
	repo.RenameServiceServiceInstance = instance
	repo.RenameServiceNewName = newName
	return
}
