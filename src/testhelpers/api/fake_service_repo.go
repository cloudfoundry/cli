package api

import (
	realApi "cf/api"
	"cf/errors"
	"cf/models"
	"generic"
)

type FakeServiceRepo struct {
	GetAllServiceOfferingsReturns struct {
		ServiceOfferings []models.ServiceOffering
		Error            errors.Error
	}

	GetServiceOfferingsForSpaceReturns struct {
		ServiceOfferings []models.ServiceOffering
		Error            errors.Error
	}
	GetServiceOfferingsForSpaceArgs struct {
		SpaceGuid string
	}

	FindServiceOfferingsForSpaceByLabelArgs struct {
		SpaceGuid string
		Name      string
	}

	FindServiceOfferingsForSpaceByLabelReturns struct {
		ServiceOfferings []models.ServiceOffering
		Error            errors.Error
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
	PurgeServiceOfferingApiResponse errors.Error

	FindServiceOfferingByLabelAndProviderName            string
	FindServiceOfferingByLabelAndProviderProvider        string
	FindServiceOfferingByLabelAndProviderServiceOffering models.ServiceOffering
	FindServiceOfferingByLabelAndProviderApiResponse     errors.Error
	FindServiceOfferingByLabelAndProviderCalled          bool

	V1ServicePlanDescription                realApi.ServicePlanDescription
	V2ServicePlanDescription                realApi.ServicePlanDescription
	FindServicePlanByDescriptionArguments   []realApi.ServicePlanDescription
	FindServicePlanByDescriptionResultGuids []string
	FindServicePlanByDescriptionResponses   []errors.Error
	findServicePlanByDescriptionCallCount   int

	ServiceInstanceCountForServicePlan int
	ServiceInstanceCountApiResponse    errors.Error

	V1GuidToMigrate                           string
	V2GuidToMigrate                           string
	MigrateServicePlanFromV1ToV2Called        bool
	MigrateServicePlanFromV1ToV2ReturnedCount int
	MigrateServicePlanFromV1ToV2Response      errors.Error
}

func (repo *FakeServiceRepo) GetAllServiceOfferings() (models.ServiceOfferings, errors.Error) {
	return repo.GetAllServiceOfferingsReturns.ServiceOfferings, repo.GetAllServiceOfferingsReturns.Error
}

func (repo *FakeServiceRepo) GetServiceOfferingsForSpace(spaceGuid string) (models.ServiceOfferings, errors.Error) {
	repo.GetServiceOfferingsForSpaceArgs.SpaceGuid = spaceGuid
	return repo.GetServiceOfferingsForSpaceReturns.ServiceOfferings, repo.GetServiceOfferingsForSpaceReturns.Error
}

func (repo *FakeServiceRepo) FindServiceOfferingsForSpaceByLabel(spaceGuid, name string) (models.ServiceOfferings, errors.Error) {
	repo.FindServiceOfferingsForSpaceByLabelArgs.Name = name
	repo.FindServiceOfferingsForSpaceByLabelArgs.SpaceGuid = spaceGuid
	return repo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings, repo.FindServiceOfferingsForSpaceByLabelReturns.Error
}

func (repo *FakeServiceRepo) PurgeServiceOffering(offering models.ServiceOffering) (apiErr errors.Error) {
	repo.PurgedServiceOffering = offering
	repo.PurgeServiceOfferingCalled = true
	return
}

func (repo *FakeServiceRepo) FindServiceOfferingByLabelAndProvider(name, provider string) (offering models.ServiceOffering, apiErr errors.Error) {
	repo.FindServiceOfferingByLabelAndProviderCalled = true
	repo.FindServiceOfferingByLabelAndProviderName = name
	repo.FindServiceOfferingByLabelAndProviderProvider = provider
	apiErr = repo.FindServiceOfferingByLabelAndProviderApiResponse
	offering = repo.FindServiceOfferingByLabelAndProviderServiceOffering
	return
}

func (repo *FakeServiceRepo) CreateServiceInstance(name, planGuid string) (identicalAlreadyExists bool, apiErr errors.Error) {
	repo.CreateServiceInstanceName = name
	repo.CreateServiceInstancePlanGuid = planGuid
	identicalAlreadyExists = repo.CreateServiceAlreadyExists

	return
}

func (repo *FakeServiceRepo) FindInstanceByName(name string) (instance models.ServiceInstance, apiErr errors.Error) {
	repo.FindInstanceByNameName = name

	if repo.FindInstanceByNameMap != nil && repo.FindInstanceByNameMap.Has(name) {
		instance = repo.FindInstanceByNameMap.Get(name).(models.ServiceInstance)
	} else {
		instance = repo.FindInstanceByNameServiceInstance
	}

	if repo.FindInstanceByNameErr {
		apiErr = errors.NewErrorWithMessage("Error finding instance")
	}

	if repo.FindInstanceByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Service instance", name)
	}

	return
}

func (repo *FakeServiceRepo) DeleteService(instance models.ServiceInstance) (apiErr errors.Error) {
	repo.DeleteServiceServiceInstance = instance
	return
}

func (repo *FakeServiceRepo) RenameService(instance models.ServiceInstance, newName string) (apiErr errors.Error) {
	repo.RenameServiceServiceInstance = instance
	repo.RenameServiceNewName = newName
	return
}

func (repo *FakeServiceRepo) FindServicePlanByDescription(planDescription realApi.ServicePlanDescription) (planGuid string, apiErr errors.Error) {

	repo.FindServicePlanByDescriptionArguments =
		append(repo.FindServicePlanByDescriptionArguments, planDescription)

	if len(repo.FindServicePlanByDescriptionResultGuids) > repo.findServicePlanByDescriptionCallCount {
		planGuid = repo.FindServicePlanByDescriptionResultGuids[repo.findServicePlanByDescriptionCallCount]
	}
	if len(repo.FindServicePlanByDescriptionResponses) > repo.findServicePlanByDescriptionCallCount {
		apiErr = repo.FindServicePlanByDescriptionResponses[repo.findServicePlanByDescriptionCallCount]
	}
	repo.findServicePlanByDescriptionCallCount += 1
	return
}

func (repo *FakeServiceRepo) GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiErr errors.Error) {
	count = repo.ServiceInstanceCountForServicePlan
	apiErr = repo.ServiceInstanceCountApiResponse
	return
}

func (repo *FakeServiceRepo) MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiErr errors.Error) {
	repo.MigrateServicePlanFromV1ToV2Called = true
	repo.V1GuidToMigrate = v1PlanGuid
	repo.V2GuidToMigrate = v2PlanGuid
	changedCount = repo.MigrateServicePlanFromV1ToV2ReturnedCount
	apiErr = repo.MigrateServicePlanFromV1ToV2Response
	return
}
