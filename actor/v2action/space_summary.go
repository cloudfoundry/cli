package v2action

import (
	"sort"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type SecurityGroupRule struct {
	Name        string
	Description string
	Destination string
	Lifecycle   ccv2.SecurityGroupLifecycle
	Ports       string
	Protocol    string
}

type SpaceSummary struct {
	Space
	OrgName                        string
	OrgDefaultIsolationSegmentGUID string
	AppNames                       []string
	ServiceInstanceNames           []string
	SpaceQuotaName                 string
	RunningSecurityGroupNames      []string
	StagingSecurityGroupNames      []string
	SecurityGroupRules             []SecurityGroupRule
}

func (actor Actor) GetSpaceSummaryByOrganizationAndName(orgGUID string, name string, includeStagingSecurityGroupsRules bool) (SpaceSummary, Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganization(orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByOrganizationAndName(org.GUID, name)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	apps, warnings, err := actor.GetApplicationsBySpace(space.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	appNames := make([]string, len(apps))
	for i, app := range apps {
		appNames[i] = app.Name
	}
	sort.Strings(appNames)

	serviceInstances, warnings, err := actor.GetServiceInstancesBySpace(space.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	serviceInstanceNames := make([]string, len(serviceInstances))
	for i, serviceInstance := range serviceInstances {
		serviceInstanceNames[i] = serviceInstance.Name
	}
	sort.Strings(serviceInstanceNames)

	var spaceQuota SpaceQuota

	if space.SpaceQuotaDefinitionGUID != "" {
		spaceQuota, warnings, err = actor.GetSpaceQuota(space.SpaceQuotaDefinitionGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return SpaceSummary{}, allWarnings, err
		}
	}

	securityGroups, warnings, err := actor.GetSpaceRunningSecurityGroupsBySpace(space.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	var runningSecurityGroupNames []string
	var stagingSecurityGroupNames []string
	var securityGroupRules []SecurityGroupRule

	for _, securityGroup := range securityGroups {
		runningSecurityGroupNames = append(runningSecurityGroupNames, securityGroup.Name)
		securityGroupRules = append(securityGroupRules, extractSecurityGroupRules(securityGroup, ccv2.SecurityGroupLifecycleRunning)...)
	}

	sort.Strings(runningSecurityGroupNames)

	if includeStagingSecurityGroupsRules {
		securityGroups, warnings, err = actor.GetSpaceStagingSecurityGroupsBySpace(space.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return SpaceSummary{}, allWarnings, err
		}

		for _, securityGroup := range securityGroups {
			stagingSecurityGroupNames = append(stagingSecurityGroupNames, securityGroup.Name)
			securityGroupRules = append(securityGroupRules, extractSecurityGroupRules(securityGroup, ccv2.SecurityGroupLifecycleStaging)...)
		}

		sort.Strings(stagingSecurityGroupNames)
	}

	sort.Slice(securityGroupRules, func(i int, j int) bool {
		if securityGroupRules[i].Name < securityGroupRules[j].Name {
			return true
		}
		if securityGroupRules[i].Name > securityGroupRules[j].Name {
			return false
		}
		if securityGroupRules[i].Destination < securityGroupRules[j].Destination {
			return true
		}
		if securityGroupRules[i].Destination > securityGroupRules[j].Destination {
			return false
		}
		return securityGroupRules[i].Lifecycle < securityGroupRules[j].Lifecycle
	})

	spaceSummary := SpaceSummary{
		Space:   space,
		OrgName: org.Name,
		OrgDefaultIsolationSegmentGUID: org.DefaultIsolationSegmentGUID,
		AppNames:                       appNames,
		ServiceInstanceNames:           serviceInstanceNames,
		SpaceQuotaName:                 spaceQuota.Name,
		RunningSecurityGroupNames:      runningSecurityGroupNames,
		StagingSecurityGroupNames:      stagingSecurityGroupNames,
		SecurityGroupRules:             securityGroupRules,
	}

	return spaceSummary, allWarnings, nil
}
