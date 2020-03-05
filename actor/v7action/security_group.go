package v7action

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	"code.cloudfoundry.org/cli/resources"
)

type SecurityGroupSummary struct {
	Name                string
	Rules               []resources.Rule
	SecurityGroupSpaces []SecurityGroupSpace
}

type SecurityGroupSpace struct {
	SpaceName string
	OrgName   string
}

func (actor Actor) CreateSecurityGroup(name, filePath string) (Warnings, error) {
	allWarnings := Warnings{}
	bytes, err := parsePath(filePath)
	if err != nil {
		return allWarnings, err
	}

	var rules []resources.Rule
	err = json.Unmarshal(bytes, &rules)
	if err != nil {
		return allWarnings, err
	}

	securityGroup := resources.SecurityGroup{Name: name, Rules: rules}

	_, warnings, err := actor.CloudControllerClient.CreateSecurityGroup(securityGroup)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}

func (actor Actor) GetSecurityGroup(securityGroupName string) (SecurityGroupSummary, Warnings, error) {
	allWarnings := Warnings{}
	securityGroupSummary := SecurityGroupSummary{}
	securityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups(ccv3.Query{Key: ccv3.NameFilter, Values: []string{securityGroupName}})

	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return securityGroupSummary, allWarnings, err
	}
	if len(securityGroups) == 0 {
		return securityGroupSummary, allWarnings, actionerror.SecurityGroupNotFoundError{Name: securityGroupName}
	}

	securityGroupSummary.Name = securityGroupName
	securityGroupSummary.Rules = securityGroups[0].Rules

	associatedSpaceGuids := securityGroups[0].RunningSpaceGUIDs
	associatedSpaceGuids = append(associatedSpaceGuids, securityGroups[0].StagingSpaceGUIDs...)

	if len(associatedSpaceGuids) > 0 {
		spaces, includes, warnings, err := actor.CloudControllerClient.GetSpaces(ccv3.Query{
			Key:    ccv3.GUIDFilter,
			Values: associatedSpaceGuids,
		}, ccv3.Query{
			Key:    ccv3.Include,
			Values: []string{"organization"},
		})

		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return securityGroupSummary, allWarnings, err
		}

		orgsByGuid := make(map[string]ccv3.Organization)
		for _, org := range includes.Organizations {
			orgsByGuid[org.GUID] = org
		}

		var securityGroupSpaces []SecurityGroupSpace
		for _, space := range spaces {
			orgGuid := space.Relationships[constant.RelationshipTypeOrganization].GUID
			securityGroupSpaces = append(securityGroupSpaces, SecurityGroupSpace{
				SpaceName: space.Name,
				OrgName:   orgsByGuid[orgGuid].Name,
			})
		}
		securityGroupSummary.SecurityGroupSpaces = securityGroupSpaces
	} else {
		securityGroupSummary.SecurityGroupSpaces = []SecurityGroupSpace{}
	}

	return securityGroupSummary, allWarnings, nil
}

func parsePath(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
