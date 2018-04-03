package director

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ReleaseSlug struct {
	name    string
	version string
}

func NewReleaseSlug(name, version string) ReleaseSlug {
	if len(name) == 0 {
		panic("Expected release to specify non-empty name")
	}

	if len(version) == 0 {
		panic("Expected release to specify non-empty version")
	}

	return ReleaseSlug{name: name, version: version}
}

func (s ReleaseSlug) Name() string    { return s.name }
func (s ReleaseSlug) Version() string { return s.version }

func (s ReleaseSlug) String() string {
	return fmt.Sprintf("%s/%s", s.name, s.version)
}

func (s *ReleaseSlug) UnmarshalFlag(data string) error {
	slug, err := parseReleaseSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseReleaseSlug(str string) (ReleaseSlug, error) {
	pieces := strings.Split(str, "/")
	if len(pieces) != 2 {
		return ReleaseSlug{}, bosherr.Errorf(
			"Expected release '%s' to be in format 'name/version'", str)
	}

	if len(pieces[0]) == 0 {
		return ReleaseSlug{}, bosherr.Errorf(
			"Expected release '%s' to specify non-empty name", str)
	}

	if len(pieces[1]) == 0 {
		return ReleaseSlug{}, bosherr.Errorf(
			"Expected release '%s' to specify non-empty version", str)
	}

	slug := ReleaseSlug{
		name:    pieces[0],
		version: pieces[1],
	}

	return slug, nil
}
