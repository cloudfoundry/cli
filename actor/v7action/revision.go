package v7action

import "code.cloudfoundry.org/cli/resources"

func (actor Actor) GetRevisionByGUID(revisionGUID string) (resources.Revision, Warnings, error) {
	revision, warnings, err := actor.CloudControllerClient.GetRevision(revisionGUID)
	if err != nil {
		return resources.Revision{}, Warnings(warnings), err
	}

	return resources.Revision(revision), Warnings(warnings), err
}
