package manifest

import (
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Validator interface {
	Validate(Manifest) error
}

type validator struct {
	logger boshlog.Logger
}

func NewValidator(logger boshlog.Logger) Validator {
	return &validator{logger: logger}
}

func (v *validator) Validate(manifest Manifest) error {
	errs := []error{}
	releaseNames := map[string]struct{}{}
	if len(manifest.Releases) < 1 {
		errs = append(errs, bosherr.Errorf("releases must contain at least 1 release"))
	}

	for releaseIdx, release := range manifest.Releases {
		if v.isBlank(release.Name) {
			errs = append(errs, bosherr.Errorf("releases[%d].name must be provided", releaseIdx))
		}

		if _, found := releaseNames[release.Name]; found {
			errs = append(errs, bosherr.Errorf("releases[%d].name '%s' must be unique", releaseIdx, release.Name))
		}
		releaseNames[release.Name] = struct{}{}

		if v.isBlank(release.URL) {
			errs = append(errs, bosherr.Errorf("releases[%d].url must be provided", releaseIdx))
		}

		matched, err := regexp.MatchString("^(file|http|https)://", release.URL)
		if err != nil || !matched {
			errs = append(errs, bosherr.Errorf("releases[%d].url must be a valid URL (file:// or http(s)://)", releaseIdx))
		}

		if strings.HasPrefix(release.URL, "http") && v.isBlank(release.SHA1) {
			errs = append(errs, bosherr.Errorf("releases[%d].sha1 must be provided for http URL", releaseIdx))
		}
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
