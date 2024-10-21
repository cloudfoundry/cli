package resources

import "code.cloudfoundry.org/cli/v7/types"

type Metadata struct {
	Labels map[string]types.NullString `json:"labels,omitempty"`
}

type ResourceMetadata struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}
