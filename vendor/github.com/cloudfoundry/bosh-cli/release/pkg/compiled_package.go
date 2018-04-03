package pkg

import (
	"fmt"

	"os"

	"github.com/cloudfoundry/bosh-cli/crypto"
	crypto2 "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CompiledPackage struct {
	name          string
	fingerprint   string
	osVersionSlug string

	Dependencies    []*CompiledPackage // todo privatize
	dependencyNames []string

	archivePath   string
	archiveDigest string
}

func NewCompiledPackageWithoutArchive(name, fp, osVersionSlug, sha1 string, dependencyNames []string) *CompiledPackage {
	return &CompiledPackage{
		name:          name,
		fingerprint:   fp,
		osVersionSlug: osVersionSlug,
		archiveDigest: sha1,

		Dependencies:    []*CompiledPackage{},
		dependencyNames: dependencyNames,
	}
}

func NewCompiledPackageWithArchive(name, fp, osVersionSlug, path, sha1 string, dependencyNames []string) *CompiledPackage {
	return &CompiledPackage{
		name:          name,
		fingerprint:   fp,
		osVersionSlug: osVersionSlug,

		archivePath:   path,
		archiveDigest: sha1,

		Dependencies:    []*CompiledPackage{},
		dependencyNames: dependencyNames,
	}
}

func (p CompiledPackage) String() string { return p.Name() }

func (p CompiledPackage) Name() string          { return p.name }
func (p CompiledPackage) Fingerprint() string   { return p.fingerprint }
func (p CompiledPackage) OSVersionSlug() string { return p.osVersionSlug }

func (p CompiledPackage) ArchivePath() string {
	if len(p.archivePath) == 0 {
		errMsg := "Internal inconsistency: Compiled package '%s/%s' does not have archive path"
		panic(fmt.Sprintf(errMsg, p.name, p.fingerprint))
	}
	return p.archivePath
}

func (p CompiledPackage) ArchiveDigest() string { return p.archiveDigest }

func (p *CompiledPackage) AttachDependencies(compiledPkgs []*CompiledPackage) error {
	for _, pkgName := range p.dependencyNames {
		var found bool

		for _, compiledPkg := range compiledPkgs {
			if compiledPkg.Name() == pkgName {
				p.Dependencies = append(p.Dependencies, compiledPkg)
				found = true
				break
			}
		}

		if !found {
			errMsg := "Expected to find compiled package '%s' since it's a dependency of compiled package '%s'"
			return bosherr.Errorf(errMsg, pkgName, p.name)
		}
	}

	return nil
}

func (p *CompiledPackage) DependencyNames() []string { return p.dependencyNames }

func (p *CompiledPackage) Deps() []Compilable {
	var coms []Compilable
	for _, dep := range p.Dependencies {
		coms = append(coms, dep)
	}
	return coms
}

func (p *CompiledPackage) IsCompiled() bool { return true }

func (p *CompiledPackage) RehashWithCalculator(digestCalculator crypto.DigestCalculator, archiveFileReader crypto2.ArchiveDigestFilePathReader) (*CompiledPackage, error) {
	pkgFile, err := archiveFileReader.OpenFile(p.archivePath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	digest, err := crypto2.ParseMultipleDigest(p.archiveDigest)
	if err != nil {
		return nil, err
	}

	err = digest.Verify(pkgFile)
	if err != nil {
		return nil, err
	}

	sha256Archive, err := digestCalculator.Calculate(p.archivePath)

	newP := *p
	newP.archiveDigest = sha256Archive

	return &newP, err
}
