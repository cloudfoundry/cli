package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"

	biutil "github.com/cloudfoundry/bosh-cli/common/util"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	birelmanifest "github.com/cloudfoundry/bosh-cli/release/manifest"
)

type Parser interface {
	Parse(string, boshtpl.Variables, patch.Op) (Manifest, error)
}

type parser struct {
	fs        boshsys.FileSystem
	validator Validator

	logTag string
	logger boshlog.Logger
}

type manifest struct {
	Releases []birelmanifest.ReleaseRef
}

func NewParser(fs boshsys.FileSystem, logger boshlog.Logger, validator Validator) Parser {
	return &parser{
		fs:        fs,
		validator: validator,

		logTag: "releaseSetParser",
		logger: logger,
	}
}

func (p *parser) Parse(path string, vars boshtpl.Variables, op patch.Op) (Manifest, error) {
	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	tpl := boshtpl.NewTemplate(contents)

	bytes, err := tpl.Evaluate(vars, op, boshtpl.EvaluateOpts{ExpectAllKeys: true})
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Evaluating manifest")
	}

	comboManifest := manifest{}

	err = yaml.Unmarshal(bytes, &comboManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Unmarshalling release set manifest")
	}

	p.logger.Debug(p.logTag, "Parsed release set manifest: %#v", comboManifest)

	for i, releaseRef := range comboManifest.Releases {
		comboManifest.Releases[i].URL, err = biutil.AbsolutifyPath(path, releaseRef.URL, p.fs)
		if err != nil {
			return Manifest{}, bosherr.WrapErrorf(err, "Resolving release path '%s", releaseRef.URL)
		}
	}

	releaseSetManifest := Manifest{
		Releases: comboManifest.Releases,
	}

	err = p.validator.Validate(releaseSetManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Validating release set manifest")
	}

	return releaseSetManifest, nil
}
