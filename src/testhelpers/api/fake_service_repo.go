package api

import (
	"cf/models"
	"cf/net"
	"generic"
	realApi "cf/api"
)

type FakeServiceRepo struct {
	ServiceOfferings []models.ServiceOffering

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

	V1ServicePlanDescription realApi.V1ServicePlanDescription
	V2ServicePlanDescription realApi.V2ServicePlanDescription

	V1FoundGuid string
	V2FoundGuid string
	FindServicePlanToMigrateByDescriptionResponse net.ApiResponse

	ServiceInstanceCountForServicePlan int
	ServiceInstanceCountApiResponse net.ApiResponse

	V1GuidToMigrate string
	V2GuidToMigrate string
	MigrateServicePlanFromV1ToV2Called bool
	MigrateServicePlanFromV1ToV2Response net.ApiResponse
}

func (repo *FakeServiceRepo) GetServiceOfferings() (offerings models.ServiceOfferings, apiResponse net.ApiResponse) {
	offerings = repo.ServiceOfferings
	return
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

func (repo *FakeServiceRepo) FindServicePlanToMigrateByDescription(v1Description realApi.V1ServicePlanDescription, v2Description realApi.V2ServicePlanDescription) (v1PlanGuid, v2PlanGuid string, apiResponse net.ApiResponse) {
	repo.V1ServicePlanDescription = v1Description
	repo.V2ServicePlanDescription = v2Description

	v1PlanGuid = repo.V1FoundGuid
	v2PlanGuid = repo.V2FoundGuid
	apiResponse = repo.FindServicePlanToMigrateByDescriptionResponse
	return
}

func (repo *FakeServiceRepo) GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiResponse net.ApiResponse) {
	count = repo.ServiceInstanceCountForServicePlan
	apiResponse = repo.ServiceInstanceCountApiResponse
	return
}


func (repo *FakeServiceRepo) MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (apiResponse net.ApiResponse) {
	repo.MigrateServicePlanFromV1ToV2Called = true
	repo.V1GuidToMigrate = v1PlanGuid
	repo.V2GuidToMigrate = v2PlanGuid
	apiResponse = repo.MigrateServicePlanFromV1ToV2Response
	return
}
