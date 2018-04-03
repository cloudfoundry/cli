package director

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ReleaseSeriesSlug struct {
	name string
}

func NewReleaseSeriesSlug(name string) ReleaseSeriesSlug {
	if len(name) == 0 {
		panic("Expected non-empty release series name")
	}
	return ReleaseSeriesSlug{name: name}
}

func (s ReleaseSeriesSlug) Name() string   { return s.name }
func (s ReleaseSeriesSlug) String() string { return s.name }

func (s *ReleaseSeriesSlug) UnmarshalFlag(data string) error {
	slug, err := parseReleaseSeriesSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseReleaseSeriesSlug(str string) (ReleaseSeriesSlug, error) {
	if len(str) == 0 {
		return ReleaseSeriesSlug{}, bosherr.Error(
			"Expected non-empty release series name")
	}

	return ReleaseSeriesSlug{name: str}, nil
}
