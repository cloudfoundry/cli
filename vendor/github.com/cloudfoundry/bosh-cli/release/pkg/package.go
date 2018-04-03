package pkg

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cloudfoundry/bosh-cli/crypto"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
	crypto2 "github.com/cloudfoundry/bosh-utils/crypto"
)

type ByName []*Package

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

type Package struct {
	resource Resource

	Dependencies    []*Package
	dependencyNames []string

	extractedPath string
	fs            boshsys.FileSystem
}

func NewPackage(resource Resource, dependencyNames []string) *Package {
	return &Package{
		resource: resource,

		Dependencies:    []*Package{},
		dependencyNames: dependencyNames,
	}
}

func NewExtractedPackage(resource Resource, dependencyNames []string, extractedPath string, fs boshsys.FileSystem) *Package {
	return &Package{
		resource: resource,

		Dependencies:    []*Package{},
		dependencyNames: dependencyNames,

		extractedPath: extractedPath,
		fs:            fs,
	}
}

func (p Package) String() string { return p.Name() }

func (p Package) Name() string        { return p.resource.Name() }
func (p Package) Fingerprint() string { return p.resource.Fingerprint() }

func (p *Package) ArchivePath() string   { return p.resource.ArchivePath() }
func (p *Package) ArchiveDigest() string { return p.resource.ArchiveDigest() }

func (p *Package) RehashWithCalculator(calculator crypto.DigestCalculator, archiveFileReader crypto2.ArchiveDigestFilePathReader) (*Package, error) {
	newResource, err := p.resource.RehashWithCalculator(calculator, archiveFileReader)
	newPkg := *p
	newPkg.resource = newResource

	return &newPkg, err
}

func (p *Package) Build(dev, final ArchiveIndex) error { return p.resource.Build(dev, final) }
func (p *Package) Finalize(final ArchiveIndex) error   { return p.resource.Finalize(final) }

func (p *Package) AttachDependencies(packages []*Package) error {
	for _, pkgName := range p.dependencyNames {
		var found bool

		for _, pkg := range packages {
			if pkg.Name() == pkgName {
				p.Dependencies = append(p.Dependencies, pkg)
				found = true
				break
			}
		}

		if !found {
			errMsg := "Expected to find package '%s' since it's a dependency of package '%s'"
			return bosherr.Errorf(errMsg, pkgName, p.Name())
		}
	}

	return nil
}

func (p *Package) DependencyNames() []string { return p.dependencyNames }

func (p *Package) Deps() []Compilable {
	var coms []Compilable
	for _, dep := range p.Dependencies {
		coms = append(coms, dep)
	}
	return coms
}

func (p *Package) IsCompiled() bool { return false }

func (p *Package) ExtractedPath() string { return p.extractedPath }

func (p *Package) CleanUp() error {
	if p.fs != nil && len(p.extractedPath) > 0 {
		return p.fs.RemoveAll(p.extractedPath)
	}
	return nil
}
