package resources

import (
	"code.cloudfoundry.org/jsonry"
)

type SharedToSpacesListWrapper struct {
	SharedToSpaceGUIDs []string       `jsonry:"data[].guid"`
	Spaces             []Space        `jsonry:"included.spaces"`
	Organizations      []Organization `jsonry:"included.organizations"`
}

func (s *SharedToSpacesListWrapper) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, s)
}
