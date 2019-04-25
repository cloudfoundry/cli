package v7pushaction

import "code.cloudfoundry.org/cli/actor/sharedaction"

// MatchResources returns back a list of matched and unmatched resources for the provided resources.
func (actor Actor) MatchResources(resources []sharedaction.V3Resource) ([]sharedaction.V3Resource, []sharedaction.V3Resource, Warnings, error) {
	matches, warnings, err := actor.V7Actor.ResourceMatch(resources)

	mapChecksumToResource := map[string]sharedaction.V3Resource{}
	for _, resource := range matches {
		mapChecksumToResource[resource.Checksum.Value] = resource
	}

	var unmatches []sharedaction.V3Resource
	for _, resource := range resources {
		if _, ok := mapChecksumToResource[resource.Checksum.Value]; !ok {
			unmatches = append(unmatches, resource)
		}
	}

	return matches, unmatches, Warnings(warnings), err
}
