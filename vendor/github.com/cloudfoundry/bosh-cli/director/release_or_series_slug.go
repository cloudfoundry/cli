package director

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ReleaseOrSeriesSlug struct {
	name    string
	version string // optional
}

func NewReleaseOrSeriesSlug(name, version string) ReleaseOrSeriesSlug {
	if len(name) == 0 {
		panic("Expected release or series to specify non-empty name")
	}
	return ReleaseOrSeriesSlug{name: name, version: version}
}

func (s ReleaseOrSeriesSlug) Name() string    { return s.name }
func (s ReleaseOrSeriesSlug) Version() string { return s.version }

func (s ReleaseOrSeriesSlug) ReleaseSlug() (ReleaseSlug, bool) {
	if len(s.version) > 0 {
		return NewReleaseSlug(s.name, s.version), true
	}
	return ReleaseSlug{}, false
}

func (s ReleaseOrSeriesSlug) SeriesSlug() ReleaseSeriesSlug {
	return NewReleaseSeriesSlug(s.name)
}

func (s *ReleaseOrSeriesSlug) UnmarshalFlag(data string) error {
	slug, err := parseReleaseOrSeriesSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseReleaseOrSeriesSlug(str string) (ReleaseOrSeriesSlug, error) {
	pieces := strings.Split(str, "/")
	if len(pieces) != 1 && len(pieces) != 2 {
		return ReleaseOrSeriesSlug{}, bosherr.Errorf(
			"Expected release or series '%s' to be in format 'name' or 'name/version'", str)
	}

	if len(pieces[0]) == 0 {
		return ReleaseOrSeriesSlug{}, bosherr.Errorf(
			"Expected release '%s' to specify non-empty name", str)
	}

	slug := ReleaseOrSeriesSlug{name: pieces[0]}

	if len(pieces) == 2 {
		if len(pieces[1]) == 0 {
			return ReleaseOrSeriesSlug{}, bosherr.Errorf(
				"Expected release '%s' to specify non-empty version", str)
		}

		slug.version = pieces[1]
	}

	return slug, nil
}
