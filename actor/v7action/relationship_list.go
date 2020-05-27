package v7action

import (
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) ShareServiceInstanceToSpaces(serviceInstanceGUID string, spaceGUIDs []string) (resources.RelationshipList, Warnings, error) {
	relationshipList, warnings, err := actor.CloudControllerClient.ShareServiceInstanceToSpaces(serviceInstanceGUID, spaceGUIDs)
	return relationshipList, Warnings(warnings), err
}
