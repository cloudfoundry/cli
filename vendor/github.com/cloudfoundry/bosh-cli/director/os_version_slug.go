package director

import (
	"encoding/json"
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type OSVersionSlug struct {
	os      string
	version string
}

func NewOSVersionSlug(os, version string) OSVersionSlug {
	if len(os) == 0 {
		panic("Expected stemcell to specify non-empty OS")
	}

	if len(version) == 0 {
		panic("Expected stemcell to specify non-empty version")
	}

	return OSVersionSlug{os: os, version: version}
}

func (s OSVersionSlug) OS() string       { return s.os }
func (s OSVersionSlug) Version() string  { return s.version }
func (s OSVersionSlug) IsProvided() bool { return len(s.os) > 0 }

func (s OSVersionSlug) String() string {
	return fmt.Sprintf("%s/%s", s.os, s.version)
}

func (s *OSVersionSlug) UnmarshalFlag(data string) error {
	slug, err := parseOSVersionSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func (s *OSVersionSlug) UnmarshalJSON(data []byte) error {
	var str string

	err := json.Unmarshal(data, &str)
	if err != nil {
		return bosherr.Errorf("Expected '%s' to be a string", data)
	}

	slug, err := parseOSVersionSlug(str)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseOSVersionSlug(str string) (OSVersionSlug, error) {
	pieces := strings.Split(str, "/")
	if len(pieces) != 2 {
		return OSVersionSlug{}, bosherr.Errorf(
			"Expected OS '%s' to be in format 'name/version'", str)
	}

	if len(pieces[0]) == 0 {
		return OSVersionSlug{}, bosherr.Errorf(
			"Expected OS '%s' to specify non-empty name", str)
	}

	if len(pieces[1]) == 0 {
		return OSVersionSlug{}, bosherr.Errorf(
			"Expected OS '%s' to specify non-empty version", str)
	}

	slug := OSVersionSlug{
		os:      pieces[0],
		version: pieces[1],
	}

	return slug, nil
}
