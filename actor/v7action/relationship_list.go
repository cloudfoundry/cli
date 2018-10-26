package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type RelationshipList ccv3.RelationshipList

func (actor Actor) ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (RelationshipList, Warnings, error) {
	relationshipList, warnings, err := actor.CloudControllerClient.ShareServiceInstanceToSpaces(serviceInstanceGUID, spaceGUIDs)
	return RelationshipList(relationshipList), Warnings(warnings), err
}
