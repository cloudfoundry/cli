package director

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type InstanceGroupOrInstanceSlug struct {
	name      string
	indexOrID string // optional
}

type InstanceFilter struct {
	Group string `json:"group"`
	ID    string `json:"id,omitempty"`
}

func NewInstanceGroupOrInstanceSlug(name, indexOrID string) InstanceGroupOrInstanceSlug {
	if len(name) == 0 {
		panic("Expected pool or instance to specify non-empty name")
	}
	return InstanceGroupOrInstanceSlug{name: name, indexOrID: indexOrID}
}

func NewInstanceGroupOrInstanceSlugFromString(str string) (InstanceGroupOrInstanceSlug, error) {
	pieces := strings.Split(str, "/")
	if len(pieces) != 1 && len(pieces) != 2 {
		return InstanceGroupOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to be in format 'name' or 'name/id-or-index'", str)
	}

	if len(pieces[0]) == 0 {
		return InstanceGroupOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to specify non-empty name", str)
	}

	slug := InstanceGroupOrInstanceSlug{name: pieces[0]}

	if len(pieces) == 2 {
		if len(pieces[1]) == 0 {
			return InstanceGroupOrInstanceSlug{}, bosherr.Errorf(
				"Expected instance '%s' to specify non-empty ID or index", str)
		}

		slug.indexOrID = pieces[1]
	}

	return slug, nil
}

func (s *InstanceGroupOrInstanceSlug) UnmarshalFlag(data string) error {
	slug, err := NewInstanceGroupOrInstanceSlugFromString(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func (s InstanceGroupOrInstanceSlug) Name() string      { return s.name }
func (s InstanceGroupOrInstanceSlug) IndexOrID() string { return s.indexOrID }

func (s InstanceGroupOrInstanceSlug) String() string {
	if len(s.indexOrID) > 0 {
		return fmt.Sprintf("%s/%s", s.name, s.indexOrID)
	}
	return s.name
}

func (s InstanceGroupOrInstanceSlug) DirectorHash() InstanceFilter {
	return InstanceFilter{
		Group: s.name,
		ID:    s.indexOrID,
	}
}
