package director

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type InstanceSlug struct {
	name      string
	indexOrID string
}

func NewInstanceSlug(name, indexOrID string) InstanceSlug {
	if len(name) == 0 {
		panic("Expected instance to specify non-empty name")
	}

	if len(indexOrID) == 0 {
		panic("Expected instance to specify non-empty index or ID")
	}

	return InstanceSlug{name: name, indexOrID: indexOrID}
}

func (s InstanceSlug) Name() string      { return s.name }
func (s InstanceSlug) IndexOrID() string { return s.indexOrID }
func (s InstanceSlug) IsProvided() bool  { return len(s.name) > 0 }

func (s InstanceSlug) String() string {
	return fmt.Sprintf("%s/%s", s.name, s.indexOrID)
}

func (s *InstanceSlug) UnmarshalFlag(data string) error {
	slug, err := parseInstanceSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseInstanceSlug(str string) (InstanceSlug, error) {
	pieces := strings.Split(str, "/")
	if len(pieces) != 2 {
		return InstanceSlug{}, bosherr.Errorf(
			"Expected instance '%s' to be in format 'name/index-or-id'", str)
	}

	if len(pieces[0]) == 0 {
		return InstanceSlug{}, bosherr.Errorf(
			"Expected instance '%s' to specify non-empty name", str)
	}

	if len(pieces[1]) == 0 {
		return InstanceSlug{}, bosherr.Errorf(
			"Expected instance '%s' to specify non-empty index or ID", str)
	}

	slug := InstanceSlug{
		name:      pieces[0],
		indexOrID: pieces[1],
	}

	return slug, nil
}
