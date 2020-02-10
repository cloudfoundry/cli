package resources

import "code.cloudfoundry.org/cli/types"

// Metadata is used for custom tagging of API resources
type Metadata struct {
	Labels map[string]types.NullString `json:"labels,omitempty"`
}

type ResourceMetadata struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}
