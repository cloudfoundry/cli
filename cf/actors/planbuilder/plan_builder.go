package planbuilder

import (
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/models"
)

//go:generate counterfeiter . PlanBuilder

type PlanBuilder interface {
	AttachOrgsToPlans([]models.ServicePlanFields) ([]models.ServicePlanFields, error)
	AttachOrgToPlans([]models.ServicePlanFields, string) ([]models.ServicePlanFields, error)
	GetPlansForServiceForOrg(string, string) ([]models.ServicePlanFields, error)
	GetPlansForServiceWithOrgs(string) ([]models.ServicePlanFields, error)
	GetPlansForManyServicesWithOrgs([]string) ([]models.ServicePlanFields, error)
	GetPlansForService(string) ([]models.ServicePlanFields, error)
	GetPlansVisibleToOrg(string) ([]models.ServicePlanFields, error)
}

var (
	OrgToPlansVisibilityMap *map[string][]string
	PlanToOrgsVisibilityMap *map[string][]string
)

type Builder struct {
	servicePlanRepo           api.ServicePlanRepository
	servicePlanVisibilityRepo api.ServicePlanVisibilityRepository
	orgRepo                   organizations.OrganizationRepository
}

func NewBuilder(plan api.ServicePlanRepository, vis api.ServicePlanVisibilityRepository, org organizations.OrganizationRepository) Builder {
	return Builder{
		servicePlanRepo:           plan,
		servicePlanVisibilityRepo: vis,
		orgRepo:                   org,
	}
}

func (builder Builder) AttachOrgToPlans(plans []models.ServicePlanFields, orgName string) ([]models.ServicePlanFields, error) {
	visMap, err := builder.buildPlanToOrgVisibilityMap(orgName)
	if err != nil {
		return nil, err
	}
	for planIndex := range plans {
		plan := &plans[planIndex]
		plan.OrgNames = visMap[plan.GUID]
	}

	return plans, nil
}

func (builder Builder) AttachOrgsToPlans(plans []models.ServicePlanFields) ([]models.ServicePlanFields, error) {
	visMap, err := builder.buildPlanToOrgsVisibilityMap()
	if err != nil {
		return nil, err
	}
	for planIndex := range plans {
		plan := &plans[planIndex]
		plan.OrgNames = visMap[plan.GUID]
	}

	return plans, nil
}

func (builder Builder) GetPlansForServiceForOrg(serviceGUID string, orgName string) ([]models.ServicePlanFields, error) {
	plans, err := builder.servicePlanRepo.Search(map[string]string{"service_guid": serviceGUID})
	if err != nil {
		return nil, err
	}

	plans, err = builder.AttachOrgToPlans(plans, orgName)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (builder Builder) GetPlansForService(serviceGUID string) ([]models.ServicePlanFields, error) {
	plans, err := builder.servicePlanRepo.Search(map[string]string{"service_guid": serviceGUID})
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (builder Builder) GetPlansForServiceWithOrgs(serviceGUID string) ([]models.ServicePlanFields, error) {
	plans, err := builder.GetPlansForService(serviceGUID)
	if err != nil {
		return nil, err
	}

	plans, err = builder.AttachOrgsToPlans(plans)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (builder Builder) GetPlansForManyServicesWithOrgs(serviceGUIDs []string) ([]models.ServicePlanFields, error) {
	plans, err := builder.servicePlanRepo.ListPlansFromManyServices(serviceGUIDs)
	if err != nil {
		return nil, err
	}

	plans, err = builder.AttachOrgsToPlans(plans)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (builder Builder) GetPlansVisibleToOrg(orgName string) ([]models.ServicePlanFields, error) {
	var plansToReturn []models.ServicePlanFields
	allPlans, err := builder.servicePlanRepo.Search(nil)

	planToOrgsVisMap, err := builder.buildPlanToOrgsVisibilityMap()
	if err != nil {
		return nil, err
	}

	orgToPlansVisMap := builder.buildOrgToPlansVisibilityMap(planToOrgsVisMap)

	filterOrgPlans := orgToPlansVisMap[orgName]

	for _, plan := range allPlans {
		if builder.containsGUID(filterOrgPlans, plan.GUID) {
			plan.OrgNames = planToOrgsVisMap[plan.GUID]
			plansToReturn = append(plansToReturn, plan)
		} else if plan.Public {
			plansToReturn = append(plansToReturn, plan)
		}
	}

	return plansToReturn, nil
}

func (builder Builder) containsGUID(guidSlice []string, guid string) bool {
	for _, g := range guidSlice {
		if g == guid {
			return true
		}
	}
	return false
}

func (builder Builder) buildPlanToOrgVisibilityMap(orgName string) (map[string][]string, error) {
	// Since this map doesn't ever change, we memoize it for performance
	orgLookup := make(map[string]string)

	org, err := builder.orgRepo.FindByName(orgName)
	if err != nil {
		return nil, err
	}
	orgLookup[org.GUID] = org.Name

	visibilities, err := builder.servicePlanVisibilityRepo.List()
	if err != nil {
		return nil, err
	}

	visMap := make(map[string][]string)
	for _, vis := range visibilities {
		if _, exists := orgLookup[vis.OrganizationGUID]; exists {
			visMap[vis.ServicePlanGUID] = append(visMap[vis.ServicePlanGUID], orgLookup[vis.OrganizationGUID])
		}
	}

	return visMap, nil
}

func (builder Builder) buildPlanToOrgsVisibilityMap() (map[string][]string, error) {
	// Since this map doesn't ever change, we memoize it for performance
	if PlanToOrgsVisibilityMap == nil {
		orgLookup := make(map[string]string)

		visibilities, err := builder.servicePlanVisibilityRepo.List()
		if err != nil {
			return nil, err
		}

		orgGUIDs := builder.getUniqueOrgGUIDsFromVisibilities(visibilities)

		orgs, err := builder.orgRepo.GetManyOrgsByGUID(orgGUIDs)
		if err != nil {
			return nil, err
		}

		for _, org := range orgs {
			orgLookup[org.GUID] = org.Name
		}

		visMap := make(map[string][]string)
		for _, vis := range visibilities {
			visMap[vis.ServicePlanGUID] = append(visMap[vis.ServicePlanGUID], orgLookup[vis.OrganizationGUID])
		}

		PlanToOrgsVisibilityMap = &visMap
	}

	return *PlanToOrgsVisibilityMap, nil
}

func (builder Builder) getUniqueOrgGUIDsFromVisibilities(visibilities []models.ServicePlanVisibilityFields) (orgGUIDs []string) {
	for _, visibility := range visibilities {
		found := false
		for _, orgGUID := range orgGUIDs {
			if orgGUID == visibility.OrganizationGUID {
				found = true
				break
			}
		}
		if !found {
			orgGUIDs = append(orgGUIDs, visibility.OrganizationGUID)
		}
	}
	return
}

func (builder Builder) buildOrgToPlansVisibilityMap(planToOrgsMap map[string][]string) map[string][]string {
	if OrgToPlansVisibilityMap == nil {
		visMap := make(map[string][]string)
		for planGUID, orgNames := range planToOrgsMap {
			for _, orgName := range orgNames {
				visMap[orgName] = append(visMap[orgName], planGUID)
			}
		}
		OrgToPlansVisibilityMap = &visMap
	}

	return *OrgToPlansVisibilityMap
}
