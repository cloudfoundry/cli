package job

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cloudfoundry/bosh-cli/crypto"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
	crypto2 "github.com/cloudfoundry/bosh-utils/crypto"
)

type ByName []*Job

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

type Job struct {
	resource Resource

	Templates    map[string]string
	PackageNames []string
	Packages     []boshpkg.Compilable
	Properties   map[string]PropertyDefinition

	extractedPath string
	fs            boshsys.FileSystem
}

type PropertyDefinition struct {
	Description string
	Default     biproperty.Property
}

func NewJob(resource Resource) *Job {
	return &Job{resource: resource}
}

func NewExtractedJob(resource Resource, extractedPath string, fs boshsys.FileSystem) *Job {
	return &Job{resource: resource, extractedPath: extractedPath, fs: fs}
}

func (j Job) Name() string        { return j.resource.Name() }
func (j Job) Fingerprint() string { return j.resource.Fingerprint() }

func (j *Job) ArchivePath() string   { return j.resource.ArchivePath() }
func (j *Job) ArchiveDigest() string { return j.resource.ArchiveDigest() }

func (j *Job) Build(dev, final ArchiveIndex) error { return j.resource.Build(dev, final) }
func (j *Job) Finalize(final ArchiveIndex) error   { return j.resource.Finalize(final) }

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}
	return "", false
}

// AttachPackages is left for testing convenience
func (j *Job) AttachPackages(packages []*boshpkg.Package) error {
	var coms []boshpkg.Compilable

	for _, pkg := range packages {
		coms = append(coms, pkg)
	}

	return j.AttachCompilablePackages(coms)
}

func (j *Job) RehashWithCalculator(calculator crypto.DigestCalculator, archiveFilePathReader crypto2.ArchiveDigestFilePathReader) (*Job, error) {
	newResource, err := j.resource.RehashWithCalculator(calculator, archiveFilePathReader)

	return &Job{
		resource:     newResource,
		Templates:    j.Templates,
		PackageNames: j.PackageNames,
		Packages:     j.Packages,
		Properties:   j.Properties,

		extractedPath: j.extractedPath,
		fs:            j.fs,
	}, err
}

func (j *Job) AttachCompilablePackages(packages []boshpkg.Compilable) error {
	for _, pkgName := range j.PackageNames {
		var found bool

		for _, pkg := range packages {
			if pkg.Name() == pkgName {
				j.Packages = append(j.Packages, pkg)
				found = true
				break
			}
		}

		if !found {
			errMsg := "Expected to find package '%s' since it's a dependency of job '%s'"
			return bosherr.Errorf(errMsg, pkgName, j.Name())
		}
	}

	return nil
}

func (j Job) ExtractedPath() string { return j.extractedPath }

func (j Job) CleanUp() error {
	if j.fs != nil && len(j.extractedPath) > 0 {
		return j.fs.RemoveAll(j.extractedPath)
	}
	return nil
}
