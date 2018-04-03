package director

import (
	"encoding/json"
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type StemcellSlug struct {
	name    string
	version string
}

func NewStemcellSlug(name, version string) StemcellSlug {
	if len(name) == 0 {
		panic("Expected stemcell to specify non-empty name")
	}

	if len(version) == 0 {
		panic("Expected stemcell to specify non-empty version")
	}

	return StemcellSlug{name: name, version: version}
}

func (s StemcellSlug) Name() string    { return s.name }
func (s StemcellSlug) Version() string { return s.version }

func (s StemcellSlug) String() string {
	return fmt.Sprintf("%s/%s", s.name, s.version)
}

func (s *StemcellSlug) UnmarshalFlag(data string) error {
	slug, err := parseStemcellSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func (s *StemcellSlug) UnmarshalJSON(data []byte) error {
	var str string

	err := json.Unmarshal(data, &str)
	if err != nil {
		return bosherr.Errorf("Expected '%s' to be a string", data)
	}

	slug, err := parseStemcellSlug(str)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseStemcellSlug(str string) (StemcellSlug, error) {
	pieces := strings.Split(str, "/")
	if len(pieces) != 2 {
		return StemcellSlug{}, bosherr.Errorf(
			"Expected stemcell '%s' to be in format 'name/version'", str)
	}

	if len(pieces[0]) == 0 {
		return StemcellSlug{}, bosherr.Errorf(
			"Expected stemcell '%s' to specify non-empty name", str)
	}

	if len(pieces[1]) == 0 {
		return StemcellSlug{}, bosherr.Errorf(
			"Expected stemcell '%s' to specify non-empty version", str)
	}

	slug := StemcellSlug{
		name:    pieces[0],
		version: pieces[1],
	}

	return slug, nil
}
