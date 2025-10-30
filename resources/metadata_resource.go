package resources

import "code.cloudfoundry.org/cli/v8/types"

type Metadata struct {
	Annotations map[string]types.NullString `json:"annotations,omitempty"`
	Labels      map[string]types.NullString `json:"labels,omitempty"`
}

type ResourceMetadata struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}
