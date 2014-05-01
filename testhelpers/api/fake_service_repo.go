package api

import (
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"
)

type FakeServiceRepo struct {
	GetAllServiceOfferingsReturns struct {
		ServiceOfferings []models.ServiceOffering
		Error            error
	}

	GetServiceOfferingsForSpaceReturns struct {
		ServiceOfferings []models.ServiceOffering
		Error            error
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
		Error            error
	}

	CreateServiceInstanceArgs struct {
		Name     string
		PlanGuid string
	}
	CreateServiceInstanceReturns struct {
		Error error
	}

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
	PurgeServiceOfferingApiResponse error

	FindServiceOfferingByLabelAndProviderName            string
	FindServiceOfferingByLabelAndProviderProvider        string
	FindServiceOfferingByLabelAndProviderServiceOffering models.ServiceOffering
	FindServiceOfferingByLabelAndProviderApiResponse     error
	FindServiceOfferingByLabelAndProviderCalled          bool

	V1ServicePlanDescription                resources.ServicePlanDescription
	V2ServicePlanDescription                resources.ServicePlanDescription
	FindServicePlanByDescriptionArguments   []resources.ServicePlanDescription
	FindServicePlanByDescriptionResultGuids []string
	FindServicePlanByDescriptionResponses   []error
	findServicePlanByDescriptionCallCount   int

	ServiceInstanceCountForServicePlan int
	ServiceInstanceCountApiResponse    error

	V1GuidToMigrate                           string
	V2GuidToMigrate                           string
	MigrateServicePlanFromV1ToV2Called        bool
	MigrateServicePlanFromV1ToV2ReturnedCount int
	MigrateServicePlanFromV1ToV2Response      error
}

func (repo *FakeServiceRepo) GetAllServiceOfferings() (models.ServiceOfferings, error) {
	return repo.GetAllServiceOfferingsReturns.ServiceOfferings, repo.GetAllServiceOfferingsReturns.Error
}

func (repo *FakeServiceRepo) GetServiceOfferingsForSpace(spaceGuid string) (models.ServiceOfferings, error) {
	repo.GetServiceOfferingsForSpaceArgs.SpaceGuid = spaceGuid
	return repo.GetServiceOfferingsForSpaceReturns.ServiceOfferings, repo.GetServiceOfferingsForSpaceReturns.Error
}

func (repo *FakeServiceRepo) FindServiceOfferingsForSpaceByLabel(spaceGuid, name string) (models.ServiceOfferings, error) {
	repo.FindServiceOfferingsForSpaceByLabelArgs.Name = name
	repo.FindServiceOfferingsForSpaceByLabelArgs.SpaceGuid = spaceGuid
	return repo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings, repo.FindServiceOfferingsForSpaceByLabelReturns.Error
}

func (repo *FakeServiceRepo) PurgeServiceOffering(offering models.ServiceOffering) (apiErr error) {
	repo.PurgedServiceOffering = offering
	repo.PurgeServiceOfferingCalled = true
	return
}

func (repo *FakeServiceRepo) FindServiceOfferingByLabelAndProvider(name, provider string) (offering models.ServiceOffering, apiErr error) {
	repo.FindServiceOfferingByLabelAndProviderCalled = true
	repo.FindServiceOfferingByLabelAndProviderName = name
	repo.FindServiceOfferingByLabelAndProviderProvider = provider
	apiErr = repo.FindServiceOfferingByLabelAndProviderApiResponse
	offering = repo.FindServiceOfferingByLabelAndProviderServiceOffering
	return
}

func (repo *FakeServiceRepo) CreateServiceInstance(name, planGuid string) (apiErr error) {
	repo.CreateServiceInstanceArgs.Name = name
	repo.CreateServiceInstanceArgs.PlanGuid = planGuid

	return repo.CreateServiceInstanceReturns.Error
}

func (repo *FakeServiceRepo) FindInstanceByName(name string) (instance models.ServiceInstance, apiErr error) {
	repo.FindInstanceByNameName = name

	if repo.FindInstanceByNameMap != nil && repo.FindInstanceByNameMap.Has(name) {
		instance = repo.FindInstanceByNameMap.Get(name).(models.ServiceInstance)
	} else {
		instance = repo.FindInstanceByNameServiceInstance
	}

	if repo.FindInstanceByNameErr {
		apiErr = errors.New("Error finding instance")
	}

	if repo.FindInstanceByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Service instance", name)
	}

	return
}

func (repo *FakeServiceRepo) DeleteService(instance models.ServiceInstance) (apiErr error) {
	repo.DeleteServiceServiceInstance = instance
	return
}

func (repo *FakeServiceRepo) RenameService(instance models.ServiceInstance, newName string) (apiErr error) {
	repo.RenameServiceServiceInstance = instance
	repo.RenameServiceNewName = newName
	return
}

func (repo *FakeServiceRepo) FindServicePlanByDescription(planDescription resources.ServicePlanDescription) (planGuid string, apiErr error) {

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

func (repo *FakeServiceRepo) GetServiceInstanceCountForServicePlan(v1PlanGuid string) (count int, apiErr error) {
	count = repo.ServiceInstanceCountForServicePlan
	apiErr = repo.ServiceInstanceCountApiResponse
	return
}

func (repo *FakeServiceRepo) MigrateServicePlanFromV1ToV2(v1PlanGuid, v2PlanGuid string) (changedCount int, apiErr error) {
	repo.MigrateServicePlanFromV1ToV2Called = true
	repo.V1GuidToMigrate = v1PlanGuid
	repo.V2GuidToMigrate = v2PlanGuid
	changedCount = repo.MigrateServicePlanFromV1ToV2ReturnedCount
	apiErr = repo.MigrateServicePlanFromV1ToV2Response
	return
}
