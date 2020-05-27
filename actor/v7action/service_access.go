package v7action

import (
	"errors"
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServicePlanWithSpaceAndOrganization ccv3.ServicePlanWithSpaceAndOrganization

type ServicePlanAccess struct {
	BrokerName          string
	ServiceOfferingName string
	ServicePlanName     string
	VisibilityType      ccv3.VisibilityType
	VisibilityDetails   []string
}

type SkippedPlans []string

const (
	visibleAdmin        = ccv3.VisibilityType("admin")
	visiblePublic       = ccv3.VisibilityType("public")
	visibleOrganization = ccv3.VisibilityType("organization")
	visibleSpace        = ccv3.VisibilityType("space")
)

type offeringDetails struct{ offeringName, brokerName string }

func (actor *Actor) GetServiceAccess(offeringName, brokerName, orgName string) ([]ServicePlanAccess, Warnings, error) {
	var orgGUID string
	var allWarnings Warnings
	if orgName != "" {
		org, orgWarnings, err := actor.GetOrganizationByName(orgName)
		if err != nil {
			return nil, orgWarnings, err
		}
		allWarnings = append(allWarnings, orgWarnings...)

		orgGUID = org.GUID
	}

	plansQuery := buildPlansFilterForGet(offeringName, brokerName, orgGUID)

	offerings, offeringsWarnings, err := actor.getServiceOfferings(offeringName, brokerName)
	allWarnings = append(allWarnings, offeringsWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	plans, plansWarnings, err := actor.CloudControllerClient.GetServicePlansWithSpaceAndOrganization(plansQuery...)
	allWarnings = append(allWarnings, plansWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var result []ServicePlanAccess
	for _, plan := range plans {
		if offering, ok := offerings[plan.ServiceOfferingGUID]; ok {
			visibilityDetails, warnings, err := actor.getServicePlanVisibilityDetails(ServicePlanWithSpaceAndOrganization(plan))
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return nil, allWarnings, err
			}

			result = append(result, ServicePlanAccess{
				ServicePlanName:     plan.Name,
				VisibilityType:      plan.VisibilityType,
				VisibilityDetails:   visibilityDetails,
				ServiceOfferingName: offering.offeringName,
				BrokerName:          offering.brokerName,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].BrokerName != result[j].BrokerName {
			return result[i].BrokerName < result[j].BrokerName
		}
		if result[i].ServiceOfferingName != result[j].ServiceOfferingName {
			return result[i].ServiceOfferingName < result[j].ServiceOfferingName
		}
		return result[i].ServicePlanName < result[j].ServicePlanName
	})

	return result, allWarnings, err
}

func (actor *Actor) EnableServiceAccess(offeringName, brokerName, orgName, planName string) (SkippedPlans, Warnings, error) {
	var allWarnings Warnings

	offering, offeringWarnings, err := actor.CloudControllerClient.GetServiceOfferingByNameAndBroker(offeringName, brokerName)
	allWarnings = append(allWarnings, offeringWarnings...)
	if err != nil {
		return nil, allWarnings, actionerror.EnrichAPIErrors(err)
	}

	plansQuery := buildPlansFilterForUpdate(offering.GUID, planName)
	plans, planWarnings, err := actor.CloudControllerClient.GetServicePlans(plansQuery...)
	allWarnings = append(allWarnings, planWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	if len(plans) == 0 {
		return nil, allWarnings, actionerror.ServicePlanNotFoundError{
			PlanName:     planName,
			OfferingName: offeringName,
		}
	}

	if offeringIsSpaceScoped(plans) {
		return nil, allWarnings, actionerror.ServicePlanVisibilityTypeError{}
	}

	visibility := ccv3.ServicePlanVisibility{Type: visiblePublic}
	if orgName != "" {
		org, orgWarnings, err := actor.GetOrganizationByName(orgName)
		allWarnings = append(allWarnings, orgWarnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		visibility.Type = visibleOrganization
		visibility.Organizations = []ccv3.VisibilityDetail{{GUID: org.GUID}}
	}

	var skipped SkippedPlans
	for _, plan := range plans {
		if plan.VisibilityType == visiblePublic && visibility.Type == visibleOrganization {
			skipped = append(skipped, plan.Name)
			continue
		}

		_, visibilityWarnings, err := actor.CloudControllerClient.UpdateServicePlanVisibility(
			plan.GUID,
			visibility,
		)
		allWarnings = append(allWarnings, visibilityWarnings...)
		if err != nil {
			return nil, allWarnings, err
		}
	}

	return skipped, allWarnings, nil
}

func (actor *Actor) DisableServiceAccess(offeringName, brokerName, orgName, planName string) (SkippedPlans, Warnings, error) {
	var allWarnings Warnings

	offering, offeringWarnings, err := actor.CloudControllerClient.GetServiceOfferingByNameAndBroker(offeringName, brokerName)
	allWarnings = append(allWarnings, offeringWarnings...)
	if err != nil {
		return SkippedPlans{}, allWarnings, actionerror.EnrichAPIErrors(err)
	}

	plansQuery := buildPlansFilterForUpdate(offering.GUID, planName)
	plans, planWarnings, err := actor.CloudControllerClient.GetServicePlans(plansQuery...)
	allWarnings = append(allWarnings, planWarnings...)
	if err != nil {
		return SkippedPlans{}, allWarnings, err
	}

	if len(plans) == 0 {
		return SkippedPlans{}, allWarnings, actionerror.ServicePlanNotFoundError{
			PlanName:     planName,
			OfferingName: offeringName,
		}
	}

	if offeringIsSpaceScoped(plans) {
		return SkippedPlans{}, allWarnings, actionerror.ServicePlanVisibilityTypeError{}
	}

	var (
		disableWarnings Warnings
		skipped         SkippedPlans
	)

	if orgName != "" {
		skipped, disableWarnings, err = actor.disableOrganizationServiceAccess(plans, orgName)
	} else {
		skipped, disableWarnings, err = actor.disableAllServiceAccess(plans)
	}

	allWarnings = append(allWarnings, disableWarnings...)
	return skipped, allWarnings, err
}

func (actor *Actor) disableAllServiceAccess(plans []ccv3.ServicePlan) (SkippedPlans, Warnings, error) {
	var (
		allWarnings Warnings
		skipped     SkippedPlans
	)

	visibility := ccv3.ServicePlanVisibility{Type: visibleAdmin}
	for _, plan := range plans {
		if plan.VisibilityType == visibleAdmin {
			skipped = append(skipped, plan.Name)
			continue
		}

		_, visibilityWarnings, err := actor.CloudControllerClient.UpdateServicePlanVisibility(
			plan.GUID,
			visibility,
		)
		allWarnings = append(allWarnings, visibilityWarnings...)
		if err != nil {
			return skipped, allWarnings, err
		}
	}
	return skipped, allWarnings, nil
}

func (actor *Actor) disableOrganizationServiceAccess(plans []ccv3.ServicePlan, orgName string) (SkippedPlans, Warnings, error) {
	var allWarnings Warnings

	org, orgWarnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, orgWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	for _, plan := range plans {
		if plan.VisibilityType == visiblePublic {
			return nil, allWarnings, errors.New("Cannot remove organization level access for public plans.")
		}
	}

	var skipped SkippedPlans
	for _, plan := range plans {
		if plan.VisibilityType == visibleAdmin {
			skipped = append(skipped, plan.Name)
			continue
		}

		deleteWarnings, err := actor.CloudControllerClient.DeleteServicePlanVisibility(plan.GUID, org.GUID)
		allWarnings = append(allWarnings, deleteWarnings...)
		if err != nil {
			return skipped, allWarnings, err
		}
	}

	return skipped, allWarnings, nil
}

func (actor *Actor) getServiceOfferings(service, broker string) (map[string]offeringDetails, Warnings, error) {
	var offeringsQuery []ccv3.Query

	if broker != "" {
		offeringsQuery = append(offeringsQuery, ccv3.Query{
			Key:    ccv3.ServiceBrokerNamesFilter,
			Values: []string{broker},
		})
	}

	if service != "" {
		offeringsQuery = append(offeringsQuery, ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{service},
		})
	}

	serviceOfferings, warnings, err := actor.CloudControllerClient.GetServiceOfferings(offeringsQuery...)
	if err != nil {
		return nil, Warnings(warnings), err
	}
	if len(serviceOfferings) == 0 && len(offeringsQuery) > 0 {
		return nil, Warnings(warnings), actionerror.ServiceNotFoundError{Name: service, Broker: broker}
	}

	offerings := make(map[string]offeringDetails)
	for _, o := range serviceOfferings {
		offerings[o.GUID] = offeringDetails{
			offeringName: o.Name,
			brokerName:   o.ServiceBrokerName,
		}
	}
	return offerings, Warnings(warnings), err
}

func (actor *Actor) getServicePlanVisibilityDetails(plan ServicePlanWithSpaceAndOrganization) (names []string, warnings Warnings, err error) {
	if plan.VisibilityType == visibleOrganization {
		result, vwarn, err := actor.CloudControllerClient.GetServicePlanVisibility(plan.GUID)
		warnings = Warnings(vwarn)
		if err != nil {
			return nil, warnings, err
		}

		for _, organization := range result.Organizations {
			names = append(names, organization.Name)
		}
	}

	if plan.VisibilityType == visibleSpace {
		names = []string{fmt.Sprintf("%s (org: %s)", plan.SpaceName, plan.OrganizationName)}
	}

	return names, warnings, nil
}

func buildPlansFilterForGet(offeringName, brokerName, orgGUID string) (query []ccv3.Query) {
	if offeringName != "" {
		query = append(query, ccv3.Query{
			Key:    ccv3.ServiceOfferingNamesFilter,
			Values: []string{offeringName},
		})
	}

	if brokerName != "" {
		query = append(query, ccv3.Query{
			Key:    ccv3.ServiceBrokerNamesFilter,
			Values: []string{brokerName},
		})
	}

	if orgGUID != "" {
		query = append(query, ccv3.Query{
			Key:    ccv3.OrganizationGUIDFilter,
			Values: []string{orgGUID},
		})
	}

	return query
}

func buildPlansFilterForUpdate(offeringGUID, planName string) []ccv3.Query {
	plansQuery := []ccv3.Query{{
		Key:    ccv3.ServiceOfferingGUIDsFilter,
		Values: []string{offeringGUID},
	}}

	if planName != "" {
		plansQuery = append(plansQuery, ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{planName},
		})
	}

	return plansQuery
}

func offeringIsSpaceScoped(plans []ccv3.ServicePlan) bool {
	// All plans from a space scoped offering will have the same visibility type
	return plans[0].VisibilityType == visibleSpace
}
