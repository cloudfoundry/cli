package v7action

import (
	"sort"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServicePlanAccess struct {
	BrokerName          string
	ServiceOfferingName string
	ServicePlanName     string
	VisibilityType      string
	VisibilityDetails   []string
}

type offeringDetails struct{ offeringName, brokerName string }

func (actor *Actor) GetServiceAccess(broker, service, organization string) ([]ServicePlanAccess, Warnings, error) {
	plansQuery, allWarnings, err := actor.buildPlansFilter(organization, broker, service)
	if err != nil {
		return nil, allWarnings, err
	}

	offerings, offeringsWarnings, err := actor.getServiceOfferings(service, broker)
	allWarnings = append(allWarnings, offeringsWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	plans, plansWarnings, err := actor.CloudControllerClient.GetServicePlans(plansQuery...)
	allWarnings = append(allWarnings, plansWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var result []ServicePlanAccess
	for _, plan := range plans {
		if offering, ok := offerings[plan.ServiceOfferingGUID]; ok {

			visibilityDetails, warnings, err := actor.getServicePlanVisibilityDetails(ServicePlan(plan))
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

func (actor *Actor) buildPlansFilter(organization, broker, service string) ([]ccv3.Query, Warnings, error) {
	var (
		plansQuery []ccv3.Query
		warnings   Warnings
	)

	if organization != "" {
		org, orgWarnings, err := actor.GetOrganizationByName(organization)
		if err != nil {
			return nil, orgWarnings, err
		}
		warnings = orgWarnings

		plansQuery = append(plansQuery, ccv3.Query{
			Key:    ccv3.OrganizationGUIDFilter,
			Values: []string{org.GUID},
		})

	}

	if broker != "" {
		plansQuery = append(plansQuery, ccv3.Query{
			Key:    ccv3.ServiceBrokerNamesFilter,
			Values: []string{broker},
		})
	}

	if service != "" {
		plansQuery = append(plansQuery, ccv3.Query{
			Key:    ccv3.ServiceOfferingNamesFilter,
			Values: []string{service},
		})
	}

	return plansQuery, warnings, nil
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

func (actor *Actor) getServicePlanVisibilityDetails(plan ServicePlan) (names []string, warnings Warnings, err error) {
	if plan.VisibilityType == "organization" {
		result, vwarn, err := actor.CloudControllerClient.GetServicePlanVisibility(plan.GUID)
		warnings = Warnings(vwarn)
		if err != nil {
			return nil, warnings, err
		}

		for _, organization := range result.Organizations {
			names = append(names, organization.Name)
		}
	}

	return names, warnings, nil
}
