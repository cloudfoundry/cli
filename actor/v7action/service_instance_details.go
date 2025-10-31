package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/extract"
	"code.cloudfoundry.org/cli/v9/util/railway"
)

const featureFlagServiceInstanceSharing string = "service_instance_sharing"

type ServiceInstanceBoundAppCount struct {
	OrgName       string
	SpaceName     string
	BoundAppCount int
}

type UsageSummaryWithSpaceAndOrg struct {
	SpaceName        string
	OrganizationName string
	BoundAppCount    int
}

type SharedStatus struct {
	FeatureFlagIsDisabled     bool
	OfferingDisablesSharing   bool
	IsSharedToOtherSpaces     bool
	IsSharedFromOriginalSpace bool
	UsageSummary              []UsageSummaryWithSpaceAndOrg
}

type ServiceInstanceParameters map[string]interface{}

type ServiceInstanceUpgradeState int

type ServiceInstanceUpgradeStatus struct {
	State       ServiceInstanceUpgradeState
	Description string
}

const (
	ServiceInstanceUpgradeNotSupported ServiceInstanceUpgradeState = iota
	ServiceInstanceUpgradeAvailable
	ServiceInstanceUpgradeNotAvailable
)

type ServiceInstanceDetails struct {
	resources.ServiceInstance
	SpaceName         string
	OrganizationName  string
	ServiceOffering   resources.ServiceOffering
	ServicePlan       resources.ServicePlan
	ServiceBrokerName string
	SharedStatus      SharedStatus
	UpgradeStatus     ServiceInstanceUpgradeStatus
	BoundApps         []resources.ServiceCredentialBinding
}

func (actor Actor) GetServiceInstanceDetails(serviceInstanceName string, spaceGUID string, omitApps bool) (ServiceInstanceDetails, Warnings, error) {
	var serviceInstanceDetails ServiceInstanceDetails

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstanceDetails, warnings, err = actor.getServiceInstanceDetails(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			serviceInstanceDetails.SharedStatus, warnings, err = actor.getServiceInstanceSharedStatus(serviceInstanceDetails, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			serviceInstanceDetails.UpgradeStatus, warnings, err = actor.getServiceInstanceUpgradeStatus(serviceInstanceDetails)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if !omitApps {
				serviceInstanceDetails.BoundApps, warnings, err = actor.getServiceInstanceBoundApps(serviceInstanceDetails.GUID)
			}
			return
		},
	)
	if err != nil {
		return ServiceInstanceDetails{}, Warnings(warnings), err
	}

	return serviceInstanceDetails, Warnings(warnings), nil
}

func (actor Actor) GetServiceInstanceParameters(serviceInstanceName string, spaceGUID string) (ServiceInstanceParameters, Warnings, error) {
	var serviceInstance resources.ServiceInstance
	var parameters ServiceInstanceParameters

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.getServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			parameters, warnings, err = actor.getServiceInstanceParameters(serviceInstance.GUID)
			return
		},
	)
	if err != nil {
		return ServiceInstanceParameters{}, Warnings(warnings), err
	}

	return parameters, Warnings(warnings), nil
}

func (actor Actor) getServiceInstanceDetails(serviceInstanceName string, spaceGUID string) (ServiceInstanceDetails, ccv3.Warnings, error) {
	query := []ccv3.Query{
		{
			Key:    ccv3.FieldsServicePlan,
			Values: []string{"name", "guid"},
		},
		{
			Key:    ccv3.FieldsServicePlanServiceOffering,
			Values: []string{"name", "guid", "description", "tags", "documentation_url"},
		},
		{
			Key:    ccv3.FieldsServicePlanServiceOfferingServiceBroker,
			Values: []string{"name", "guid"},
		},
		{
			Key:    ccv3.FieldsSpace,
			Values: []string{"name", "guid"},
		},
		{
			Key:    ccv3.FieldsSpaceOrganization,
			Values: []string{"name", "guid"},
		},
	}

	serviceInstance, included, warnings, err := actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID, query...)
	switch err.(type) {
	case nil:
	case ccerror.ServiceInstanceNotFoundError:
		return ServiceInstanceDetails{}, warnings, actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	default:
		return ServiceInstanceDetails{}, warnings, err
	}

	result := ServiceInstanceDetails{
		ServiceInstance:   serviceInstance,
		ServicePlan:       extractServicePlan(included),
		ServiceOffering:   extractServiceOffering(included),
		ServiceBrokerName: extract.First("Name", included.ServiceBrokers),
		SpaceName:         extract.First("Name", included.Spaces),
		OrganizationName:  extract.First("Name", included.Organizations),
	}

	return result, warnings, nil
}

func (actor Actor) getServiceInstanceParameters(serviceInstanceGUID string) (ServiceInstanceParameters, ccv3.Warnings, error) {
	params, warnings, err := actor.CloudControllerClient.GetServiceInstanceParameters(serviceInstanceGUID)

	switch err := err.(type) {
	case nil:
		return ServiceInstanceParameters(params), warnings, nil
	case ccerror.ResourceNotFoundError,
		ccerror.ServiceInstanceParametersFetchNotSupportedError:
		return ServiceInstanceParameters{}, warnings, actionerror.ServiceInstanceParamsFetchingNotSupportedError{}
	default:
		return ServiceInstanceParameters{}, warnings, err
	}
}

func (actor Actor) getServiceInstanceSharedStatus(serviceInstanceDetails ServiceInstanceDetails, targetedSpace string) (SharedStatus, ccv3.Warnings, error) {
	if serviceInstanceDetails.Type != resources.ManagedServiceInstance {
		return SharedStatus{}, nil, nil
	}

	if targetedSpace != serviceInstanceDetails.SpaceGUID {
		return SharedStatus{IsSharedFromOriginalSpace: true}, nil, nil
	}

	var (
		featureFlag             resources.FeatureFlag
		offeringDisablesSharing bool
		sharedSpaces            []ccv3.SpaceWithOrganization
		usageSummaries          []resources.ServiceInstanceUsageSummary
	)

	warnings, err := railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			featureFlag, warnings, err = actor.CloudControllerClient.GetFeatureFlag(featureFlagServiceInstanceSharing)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			offeringDisablesSharing, warnings, err = actor.getOfferingSharingDetails(serviceInstanceDetails.ServiceOffering.GUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			sharedSpaces, warnings, err = actor.CloudControllerClient.GetServiceInstanceSharedSpaces(serviceInstanceDetails.GUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if len(sharedSpaces) > 0 {
				usageSummaries, warnings, err = actor.CloudControllerClient.GetServiceInstanceUsageSummary(serviceInstanceDetails.GUID)
			}
			return
		},
	)
	if err != nil {
		return SharedStatus{}, warnings, err
	}

	sharedStatus := SharedStatus{
		IsSharedToOtherSpaces:   len(sharedSpaces) > 0,
		OfferingDisablesSharing: offeringDisablesSharing,
		FeatureFlagIsDisabled:   !featureFlag.Enabled,
		UsageSummary:            buildUsageSummary(sharedSpaces, usageSummaries),
	}

	return sharedStatus, warnings, nil
}

func (actor Actor) getOfferingSharingDetails(serviceOfferingGUID string) (bool, ccv3.Warnings, error) {
	serviceOffering, serviceOfferingWarning, err :=
		actor.CloudControllerClient.GetServiceOfferingByGUID(serviceOfferingGUID)

	switch err := err.(type) {
	case nil:
		return !serviceOffering.AllowsInstanceSharing, serviceOfferingWarning, nil
	case ccerror.ServiceOfferingNotFoundError:
		return false, serviceOfferingWarning, nil
	default:
		return false, serviceOfferingWarning, err
	}
}

func (actor Actor) getServiceInstanceUpgradeStatus(serviceInstanceDetails ServiceInstanceDetails) (ServiceInstanceUpgradeStatus, ccv3.Warnings, error) {
	if !serviceInstanceDetails.UpgradeAvailable.Value {
		if serviceInstanceDetails.MaintenanceInfoVersion == "" {
			return ServiceInstanceUpgradeStatus{State: ServiceInstanceUpgradeNotSupported}, nil, nil
		}
		return ServiceInstanceUpgradeStatus{State: ServiceInstanceUpgradeNotAvailable}, nil, nil
	}

	servicePlan, warnings, err := actor.CloudControllerClient.GetServicePlanByGUID(serviceInstanceDetails.ServicePlanGUID)
	switch err.(type) {
	case nil:
		return ServiceInstanceUpgradeStatus{
			State:       ServiceInstanceUpgradeAvailable,
			Description: servicePlan.MaintenanceInfoDescription,
		}, warnings, nil
	case ccerror.ServicePlanNotFound:
		return ServiceInstanceUpgradeStatus{
			State:       ServiceInstanceUpgradeAvailable,
			Description: "No upgrade details where found",
		}, warnings, nil
	default:
		return ServiceInstanceUpgradeStatus{}, warnings, err
	}
}

func (actor Actor) getServiceInstanceBoundApps(serviceInstanceGUID string) ([]resources.ServiceCredentialBinding, ccv3.Warnings, error) {
	return actor.CloudControllerClient.GetServiceCredentialBindings(
		ccv3.Query{Key: ccv3.Include, Values: []string{"app"}},
		ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
		ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"app"}},
	)
}

func extractServicePlan(included ccv3.IncludedResources) resources.ServicePlan {
	if len(included.ServicePlans) == 1 {
		return included.ServicePlans[0]
	}

	return resources.ServicePlan{}
}

func extractServiceOffering(included ccv3.IncludedResources) resources.ServiceOffering {
	if len(included.ServiceOfferings) == 1 {
		return included.ServiceOfferings[0]
	}

	return resources.ServiceOffering{}
}

func buildUsageSummary(sharedSpaces []ccv3.SpaceWithOrganization, usageSummaries []resources.ServiceInstanceUsageSummary) []UsageSummaryWithSpaceAndOrg {
	var spaceGUIDToNames = make(map[string]ccv3.SpaceWithOrganization)
	var sharedSpacesUsage []UsageSummaryWithSpaceAndOrg

	for _, sharedSpace := range sharedSpaces {
		spaceGUIDToNames[sharedSpace.SpaceGUID] = sharedSpace
	}
	for _, usageSummary := range usageSummaries {
		summary := UsageSummaryWithSpaceAndOrg{
			SpaceName:        spaceGUIDToNames[usageSummary.SpaceGUID].SpaceName,
			OrganizationName: spaceGUIDToNames[usageSummary.SpaceGUID].OrganizationName,
			BoundAppCount:    usageSummary.BoundAppCount,
		}
		sharedSpacesUsage = append(sharedSpacesUsage, summary)
	}
	return sharedSpacesUsage
}
