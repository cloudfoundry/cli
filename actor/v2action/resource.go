package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultFolderPermissions      = 0755
	DefaultArchiveFilePermissions = 0744
	MaxResourceMatchChunkSize     = 1000
)

var DefaultIgnoreLines = []string{
	".cfignore",
	".DS_Store",
	".git",
	".gitignore",
	".hg",
	".svn",
	"_darcs",
	"manifest.yaml",
	"manifest.yml",
}

type Resource ccv2.Resource

// ResourceMatch returns a set of matched resources and unmatched resources in
// the order they were given in allResources.
func (actor Actor) ResourceMatch(allResources []Resource) ([]Resource, []Resource, Warnings, error) {
	resourcesToSend := [][]ccv2.Resource{{}}
	var currentList, sendCount int
	for _, resource := range allResources {
		// Skip if resource is a directory, symlink, or empty file.
		if resource.Size == 0 {
			continue
		}

		resourcesToSend[currentList] = append(
			resourcesToSend[currentList],
			ccv2.Resource(resource),
		)
		sendCount++

		if len(resourcesToSend[currentList]) == MaxResourceMatchChunkSize {
			currentList++
			resourcesToSend = append(resourcesToSend, []ccv2.Resource{})
		}
	}

	log.WithFields(log.Fields{
		"total_resources":    len(allResources),
		"resources_to_match": sendCount,
		"chunks":             len(resourcesToSend),
	}).Debug("sending resource match stats")

	matchedCCResources := map[string]ccv2.Resource{}
	var allWarnings Warnings
	for _, chunk := range resourcesToSend {
		if len(chunk) == 0 {
			log.Debug("chunk size 0, stopping resource match requests")
			break
		}

		returnedResources, warnings, err := actor.CloudControllerClient.ResourceMatch(chunk)
		allWarnings = append(allWarnings, warnings...)

		if err != nil {
			log.Errorln("during resource matching", err)
			return nil, nil, allWarnings, err
		}

		for _, resource := range returnedResources {
			matchedCCResources[resource.SHA1] = resource
		}
	}
	log.WithField("matched_resource_count", len(matchedCCResources)).Debug("total number of matched resources")

	var matchedResources, unmatchedResources []Resource
	for _, resource := range allResources {
		if _, ok := matchedCCResources[resource.SHA1]; ok {
			matchedResources = append(matchedResources, resource)
		} else {
			unmatchedResources = append(unmatchedResources, resource)
		}
	}

	return matchedResources, unmatchedResources, allWarnings, nil
}

func (Actor) actorToCCResources(resources []Resource) []ccv2.Resource {
	apiResources := make([]ccv2.Resource, 0, len(resources)) // Explicitly done to prevent nils

	for _, resource := range resources {
		apiResources = append(apiResources, ccv2.Resource(resource))
	}

	return apiResources
}
