package cfnetworkingaction

import (
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
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

	srcApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(srcAppName, srcSpaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	destApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(destAppName, destSpaceGUID)
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

	currentSpaceApps, warnings, err := actor.V3Actor.GetApplicationsBySpace(spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	var srcAppGUIDs []string
	for _, app := range currentSpaceApps {
		srcAppGUIDs = append(srcAppGUIDs, app.GUID)
	}

	var v1Policies []cfnetv1.Policy
	v1Policies, err = actor.NetworkingClient.ListPolicies(srcAppGUIDs...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	v1Policies = filterPoliciesWithoutMatchingSourceGUIDs(v1Policies, srcAppGUIDs)

	destAppGUIDs := uniqueDestGUIDs(v1Policies)

	destApplications, warnings, err := actor.V3Actor.GetApplicationsByGUIDs(destAppGUIDs...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	applications := append(currentSpaceApps, destApplications...)
	spaceGUIDs := uniqueSpaceGUIDs(applications)

	spaces, warnings, err := actor.V3Actor.GetSpacesByGUIDs(spaceGUIDs...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	spaceNamesByGUID := make(map[string]string, len(spaces))
	for _, destSpace := range spaces {
		spaceNamesByGUID[destSpace.GUID] = destSpace.Name
	}

	orgGUIDs := uniqueOrgGUIDs(spaces)

	orgs, warnings, err := actor.V3Actor.GetOrganizationsByGUIDs(orgGUIDs...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	orgNamesByGUID := make(map[string]string, len(orgs))
	for _, org := range orgs {
		orgNamesByGUID[org.GUID] = org.Name
	}

	orgNamesBySpaceGUID := make(map[string]string, len(spaces))
	for _, space := range spaces {
		orgNamesBySpaceGUID[space.GUID] = orgNamesByGUID[space.OrganizationGUID]
	}

	appByGUID := map[string]v3action.Application{}
	for _, app := range applications {
		appByGUID[app.GUID] = app
	}

	var policies []Policy
	for _, v1Policy := range v1Policies {
		policies = append(policies, actor.transformPolicy(appByGUID, spaceNamesByGUID, orgNamesBySpaceGUID, v1Policy))
	}

	return policies, allWarnings, nil
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

func uniqueSpaceGUIDs(applications []v3action.Application) []string {
	var spaceGUIDs []string
	occurances := map[string]struct{}{}
	for _, app := range applications {
		if _, ok := occurances[app.SpaceGUID]; !ok {
			spaceGUIDs = append(spaceGUIDs, app.SpaceGUID)
			occurances[app.SpaceGUID] = struct{}{}
		}
	}
	return spaceGUIDs
}

func uniqueOrgGUIDs(spaces []v3action.Space) []string {
	var orgGUIDs []string
	occurances := map[string]struct{}{}
	for _, space := range spaces {
		if _, ok := occurances[space.OrganizationGUID]; !ok {
			orgGUIDs = append(orgGUIDs, space.OrganizationGUID)
			occurances[space.OrganizationGUID] = struct{}{}
		}
	}
	return orgGUIDs
}

func uniqueDestGUIDs(policies []cfnetv1.Policy) []string {
	var destAppGUIDs []string
	occurances := map[string]struct{}{}
	for _, policy := range policies {
		if _, ok := occurances[policy.Destination.ID]; !ok {
			destAppGUIDs = append(destAppGUIDs, policy.Destination.ID)
			occurances[policy.Destination.ID] = struct{}{}
		}
	}
	return destAppGUIDs
}

func (actor Actor) NetworkPoliciesBySpaceAndAppName(spaceGUID string, srcAppName string) ([]Policy, Warnings, error) {
	var allWarnings Warnings

	srcApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(srcAppName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	var v1Policies []cfnetv1.Policy
	v1Policies, err = actor.NetworkingClient.ListPolicies(srcApp.GUID)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	v1Policies = filterPoliciesWithoutMatchingSourceGUIDs(v1Policies, []string{srcApp.GUID})

	destAppGUIDs := uniqueDestGUIDs(v1Policies)

	destApplications, warnings, err := actor.V3Actor.GetApplicationsByGUIDs(destAppGUIDs...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	applications := append(destApplications, srcApp)
	spaceGUIDs := uniqueSpaceGUIDs(applications)

	spaces, warnings, err := actor.V3Actor.GetSpacesByGUIDs(spaceGUIDs...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	spaceNamesByGUID := make(map[string]string, len(spaces))
	for _, destSpace := range spaces {
		spaceNamesByGUID[destSpace.GUID] = destSpace.Name
	}

	for _, destSpace := range spaces {
		spaceNamesByGUID[destSpace.GUID] = destSpace.Name
	}

	orgGUIDs := uniqueOrgGUIDs(spaces)

	orgs, warnings, err := actor.V3Actor.GetOrganizationsByGUIDs(orgGUIDs...)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	orgNamesByGUID := make(map[string]string, len(orgs))
	for _, org := range orgs {
		orgNamesByGUID[org.GUID] = org.Name
	}

	orgNamesBySpaceGUID := make(map[string]string, len(spaces))
	for _, space := range spaces {
		orgNamesBySpaceGUID[space.GUID] = orgNamesByGUID[space.OrganizationGUID]
	}

	appByGUID := map[string]v3action.Application{}
	for _, app := range applications {
		appByGUID[app.GUID] = app
	}

	var policies []Policy
	for _, v1Policy := range v1Policies {
		policies = append(policies, actor.transformPolicy(appByGUID, spaceNamesByGUID, orgNamesBySpaceGUID, v1Policy))
	}

	return policies, allWarnings, nil
}

func (actor Actor) RemoveNetworkPolicy(srcSpaceGUID, srcAppName, destSpaceGUID, destAppName, protocol string, startPort, endPort int) (Warnings, error) {
	var allWarnings Warnings

	srcApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(srcAppName, srcSpaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	destApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(destAppName, destSpaceGUID)
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

func (Actor) transformPolicy(appByGuid map[string]v3action.Application, spaceNamesByGUID, orgNamesBySpaceGUID map[string]string, v1Policy cfnetv1.Policy) Policy {
	dst := appByGuid[v1Policy.Destination.ID]
	return Policy{
		SourceName:           appByGuid[v1Policy.Source.ID].Name,
		DestinationName:      dst.Name,
		Protocol:             string(v1Policy.Destination.Protocol),
		StartPort:            v1Policy.Destination.Ports.Start,
		EndPort:              v1Policy.Destination.Ports.End,
		DestinationSpaceName: spaceNamesByGUID[dst.SpaceGUID],
		DestinationOrgName:   orgNamesBySpaceGUID[dst.SpaceGUID],
	}
}
