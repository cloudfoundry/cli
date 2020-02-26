package cfnetworkingaction

import (
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Policy struct {
	SourceName           string
	DestinationName      string
	Protocol             string
	DestinationSpaceName string
	DestinationOrgName   string
	StartPort            int
	EndPort              int
}

func (actor Actor) AddNetworkPolicy(srcSpaceGUID, srcAppName, destSpaceGUID, destAppName, protocol string, startPort, endPort int) (Warnings, error) {
	var allWarnings Warnings

	srcApp, warnings, err := actor.CloudControllerClient.GetApplicationByNameAndSpace(srcAppName, srcSpaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	destApp, warnings, err := actor.CloudControllerClient.GetApplicationByNameAndSpace(destAppName, destSpaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	err = actor.NetworkingClient.CreatePolicies([]cfnetv1.Policy{
		{
			Source: cfnetv1.PolicySource{
				ID: srcApp.GUID,
			},
			Destination: cfnetv1.PolicyDestination{
				ID:       destApp.GUID,
				Protocol: cfnetv1.PolicyProtocol(protocol),
				Ports: cfnetv1.Ports{
					Start: startPort,
					End:   endPort,
				},
			},
		},
	})
	return allWarnings, err
}

func (actor Actor) NetworkPoliciesBySpace(spaceGUID string) ([]Policy, Warnings, error) {
	var allWarnings Warnings

	applications, warnings, err := actor.CloudControllerClient.GetApplications(ccv3.Query{
		Key:    ccv3.SpaceGUIDFilter,
		Values: []string{spaceGUID},
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	policies, warnings, err := actor.getPoliciesForApplications(applications)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	return policies, allWarnings, nil
}

func (actor Actor) NetworkPoliciesBySpaceAndAppName(spaceGUID string, srcAppName string) ([]Policy, Warnings, error) {
	var allWarnings Warnings

	srcApp, warnings, err := actor.CloudControllerClient.GetApplicationByNameAndSpace(srcAppName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	policies, warnings, err := actor.getPoliciesForApplications([]ccv3.Application{srcApp})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	return policies, allWarnings, nil
}

func (actor Actor) RemoveNetworkPolicy(srcSpaceGUID, srcAppName, destSpaceGUID, destAppName, protocol string, startPort, endPort int) (Warnings, error) {
	var allWarnings Warnings

	srcApp, warnings, err := actor.CloudControllerClient.GetApplicationByNameAndSpace(srcAppName, srcSpaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	destApp, warnings, err := actor.CloudControllerClient.GetApplicationByNameAndSpace(destAppName, destSpaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	policyToRemove := cfnetv1.Policy{
		Source: cfnetv1.PolicySource{
			ID: srcApp.GUID,
		},
		Destination: cfnetv1.PolicyDestination{
			ID:       destApp.GUID,
			Protocol: cfnetv1.PolicyProtocol(protocol),
			Ports: cfnetv1.Ports{
				Start: startPort,
				End:   endPort,
			},
		},
	}

	v1Policies, err := actor.NetworkingClient.ListPolicies(srcApp.GUID)
	if err != nil {
		return allWarnings, err
	}

	for _, v1Policy := range v1Policies {
		if v1Policy == policyToRemove {
			return allWarnings, actor.NetworkingClient.RemovePolicies([]cfnetv1.Policy{policyToRemove})
		}
	}

	return allWarnings, actionerror.PolicyDoesNotExistError{}
}

func filterPoliciesWithoutMatchingSourceGUIDs(v1Policies []cfnetv1.Policy, srcAppGUIDs []string) []cfnetv1.Policy {
	srcGUIDsSet := map[string]struct{}{}
	for _, srcGUID := range srcAppGUIDs {
		srcGUIDsSet[srcGUID] = struct{}{}
	}

	var toReturn []cfnetv1.Policy
	for _, policy := range v1Policies {
		if _, ok := srcGUIDsSet[policy.Source.ID]; ok {
			toReturn = append(toReturn, policy)
		}
	}

	return toReturn
}

func uniqueSpaceGUIDs(applications []ccv3.Application) []string {
	var spaceGUIDs []string
	occurrences := map[string]struct{}{}
	for _, app := range applications {
		spaceGUID := app.Relationships[constant.RelationshipTypeSpace].GUID

		if _, ok := occurrences[spaceGUID]; !ok {
			spaceGUIDs = append(spaceGUIDs, spaceGUID)
			occurrences[spaceGUID] = struct{}{}
		}
	}
	return spaceGUIDs
}

func uniqueOrgGUIDs(spaces []ccv3.Space) []string {
	var orgGUIDs []string
	occurrences := map[string]struct{}{}
	for _, space := range spaces {
		orgGUID := space.Relationships[constant.RelationshipTypeOrganization].GUID

		if _, ok := occurrences[orgGUID]; !ok {
			orgGUIDs = append(orgGUIDs, orgGUID)
			occurrences[orgGUID] = struct{}{}
		}
	}
	return orgGUIDs
}

func uniqueDestGUIDs(policies []cfnetv1.Policy) []string {
	var destAppGUIDs []string
	occurrences := map[string]struct{}{}
	for _, policy := range policies {
		if _, ok := occurrences[policy.Destination.ID]; !ok {
			destAppGUIDs = append(destAppGUIDs, policy.Destination.ID)
			occurrences[policy.Destination.ID] = struct{}{}
		}
	}
	return destAppGUIDs
}

func (actor Actor) orgNamesBySpaceGUID(spaces []ccv3.Space) (map[string]string, ccv3.Warnings, error) {
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: uniqueOrgGUIDs(spaces),
	})
	if err != nil {
		return nil, warnings, err
	}

	orgNamesByGUID := make(map[string]string, len(orgs))
	for _, org := range orgs {
		orgNamesByGUID[org.GUID] = org.Name
	}

	orgNamesBySpaceGUID := make(map[string]string, len(spaces))
	for _, space := range spaces {
		orgGUID := space.Relationships[constant.RelationshipTypeOrganization].GUID
		orgNamesBySpaceGUID[space.GUID] = orgNamesByGUID[orgGUID]
	}

	return orgNamesBySpaceGUID, warnings, nil
}

func (actor Actor) getPoliciesForApplications(applications []ccv3.Application) ([]Policy, ccv3.Warnings, error) {
	var allWarnings ccv3.Warnings

	var srcAppGUIDs []string
	for _, app := range applications {
		srcAppGUIDs = append(srcAppGUIDs, app.GUID)
	}

	v1Policies, err := actor.NetworkingClient.ListPolicies(srcAppGUIDs...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	// ListPolicies will return policies with the app guids in either the source or destination.
	// It needs to be further filtered to only get policies with the app guids in the source.
	v1Policies = filterPoliciesWithoutMatchingSourceGUIDs(v1Policies, srcAppGUIDs)

	destAppGUIDs := uniqueDestGUIDs(v1Policies)

	destApplications, warnings, err := actor.CloudControllerClient.GetApplications(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: destAppGUIDs,
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	applications = append(applications, destApplications...)

	spaces, _, warnings, err := actor.CloudControllerClient.GetSpaces(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: uniqueSpaceGUIDs(applications),
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	spaceNamesByGUID := make(map[string]string, len(spaces))
	for _, destSpace := range spaces {
		spaceNamesByGUID[destSpace.GUID] = destSpace.Name
	}

	orgNamesBySpaceGUID, warnings, err := actor.orgNamesBySpaceGUID(spaces)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	appByGUID := map[string]ccv3.Application{}
	for _, app := range applications {
		appByGUID[app.GUID] = app
	}

	var policies []Policy
	for _, v1Policy := range v1Policies {
		destination := appByGUID[v1Policy.Destination.ID]
		spaceGUID := destination.Relationships[constant.RelationshipTypeSpace].GUID

		policies = append(policies, Policy{
			SourceName:           appByGUID[v1Policy.Source.ID].Name,
			DestinationName:      destination.Name,
			Protocol:             string(v1Policy.Destination.Protocol),
			StartPort:            v1Policy.Destination.Ports.Start,
			EndPort:              v1Policy.Destination.Ports.End,
			DestinationSpaceName: spaceNamesByGUID[spaceGUID],
			DestinationOrgName:   orgNamesBySpaceGUID[spaceGUID],
		})
	}

	return policies, allWarnings, nil
}
