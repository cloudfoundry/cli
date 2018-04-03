package director

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type InstanceGroupSlug struct {
	name string
}

func NewInstanceGroupSlug(name string) InstanceGroupSlug {
	if len(name) == 0 {
		panic("Expected non-empty pool name")
	}
	return InstanceGroupSlug{name: name}
}

func (s InstanceGroupSlug) Name() string   { return s.name }
func (s InstanceGroupSlug) String() string { return s.name }

func (s *InstanceGroupSlug) UnmarshalFlag(data string) error {
	slug, err := parseInstanceGroupSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseInstanceGroupSlug(str string) (InstanceGroupSlug, error) {
	if len(str) == 0 {
		return InstanceGroupSlug{}, bosherr.Error("Expected non-empty pool name")
	}

	return InstanceGroupSlug{name: str}, nil
}
