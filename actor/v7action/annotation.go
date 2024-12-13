package v7action

import (
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

func (actor *Actor) GetRevisionAnnotations(revisionGUID string) (map[string]types.NullString, Warnings, error) {
	resource, warnings, err := actor.GetRevisionByGUID(revisionGUID)
	return actor.extractAnnotations(resource.Metadata, warnings, err)
}

func (actor *Actor) extractAnnotations(metadata *resources.Metadata, warnings Warnings, err error) (map[string]types.NullString, Warnings, error) {
	var annotations map[string]types.NullString

	if err != nil {
		return annotations, warnings, err
	}
	if metadata != nil {
		annotations = metadata.Annotations
	}
	return annotations, warnings, nil
}
