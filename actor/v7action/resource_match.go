package v7action

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) ResourceMatch(resources []sharedaction.V3Resource) ([]sharedaction.V3Resource, Warnings, error) {
	resourceChunks := actor.chunkResources(resources)

	log.WithFields(log.Fields{
		"total_resources": len(resources),
		"chunks":          len(resourceChunks),
	}).Debug("sending resource match stats")

	var (
		allWarnings         Warnings
		matchedAPIResources []ccv3.Resource
	)

	for _, chunk := range resourceChunks {
		newMatchedAPIResources, warnings, err := actor.CloudControllerClient.ResourceMatch(chunk)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		matchedAPIResources = append(matchedAPIResources, newMatchedAPIResources...)
	}

	var matchedResources []sharedaction.V3Resource
	for _, resource := range matchedAPIResources {
		matchedResources = append(matchedResources, sharedaction.V3Resource(resource))
	}

	log.WithFields(log.Fields{
		"matchedResources": len(matchedResources),
	}).Debug("number of resources matched by CC")

	return matchedResources, allWarnings, nil
}

func (Actor) chunkResources(resources []sharedaction.V3Resource) [][]ccv3.Resource {
	var chunkedResources [][]ccv3.Resource
	var currentSet []ccv3.Resource

	for index, resource := range resources {
		if resource.SizeInBytes != 0 {
			currentSet = append(currentSet, ccv3.Resource(resource))
		}

		if len(currentSet) == constant.MaxNumberOfResourcesForMatching || index+1 == len(resources) {
			chunkedResources = append(chunkedResources, currentSet)
			currentSet = []ccv3.Resource{}
		}
	}
	return chunkedResources
}
