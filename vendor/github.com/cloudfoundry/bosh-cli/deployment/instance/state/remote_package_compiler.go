package state

import (
	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	biblobstore "github.com/cloudfoundry/bosh-cli/blobstore"
	birelpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	bistatepkg "github.com/cloudfoundry/bosh-cli/state/pkg"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type remotePackageCompiler struct {
	blobstore   biblobstore.Blobstore
	agentClient biagentclient.AgentClient
	packageRepo bistatepkg.CompiledPackageRepo
}

func NewRemotePackageCompiler(
	blobstore biblobstore.Blobstore,
	agentClient biagentclient.AgentClient,
	packageRepo bistatepkg.CompiledPackageRepo,
) bistatepkg.Compiler {
	return &remotePackageCompiler{
		blobstore:   blobstore,
		agentClient: agentClient,
		packageRepo: packageRepo,
	}
}

func (c *remotePackageCompiler) Compile(pkg birelpkg.Compilable) (bistatepkg.CompiledPackageRecord, bool, error) {
	var record bistatepkg.CompiledPackageRecord

	blobID, err := c.blobstore.Add(pkg.ArchivePath())
	if err != nil {
		return bistatepkg.CompiledPackageRecord{}, false, bosherr.WrapErrorf(err, "Adding release package archive '%s' to blobstore", pkg.ArchivePath())
	}

	packageSource := biagentclient.BlobRef{
		Name:        pkg.Name(),
		Version:     pkg.Fingerprint(),
		SHA1:        pkg.ArchiveDigest(),
		BlobstoreID: blobID,
	}

	var isAlreadyCompiled bool

	if !pkg.IsCompiled() {
		// Resolve dependencies from map of previously compiled packages.
		// Only install the package's immediate dependencies.
		packageDependencies := make([]biagentclient.BlobRef, len(pkg.Deps()), len(pkg.Deps()))

		for i, pkgDep := range pkg.Deps() {
			compiledPackageRecord, found, err := c.packageRepo.Find(pkgDep)
			if err != nil {
				return record, false, bosherr.WrapErrorf(
					err,
					"Finding compiled package '%s/%s' as pkgDep for '%s/%s'",
					pkgDep.Name(),
					pkgDep.Fingerprint(),
					pkg.Name(),
					pkg.Fingerprint(),
				)
			}
			if !found {
				return record, false, bosherr.Errorf(
					"Remote compilation failure: Package '%s/%s' requires package '%s/%s', but it has not been compiled",
					pkg.Name(),
					pkg.Fingerprint(),
					pkgDep.Name(),
					pkgDep.Fingerprint(),
				)
			}
			packageDependencies[i] = biagentclient.BlobRef{
				Name:        pkgDep.Name(),
				Version:     pkgDep.Fingerprint(),
				BlobstoreID: compiledPackageRecord.BlobID,
				SHA1:        compiledPackageRecord.BlobSHA1,
			}
		}

		compiledPackageRef, err := c.agentClient.CompilePackage(packageSource, packageDependencies)
		if err != nil {
			return record, false, bosherr.WrapErrorf(err, "Remotely compiling package '%s' with the agent", pkg.Name())
		}

		record = bistatepkg.CompiledPackageRecord{
			BlobID:   compiledPackageRef.BlobstoreID,
			BlobSHA1: compiledPackageRef.SHA1,
		}
	} else {
		isAlreadyCompiled = true

		record = bistatepkg.CompiledPackageRecord{
			BlobID:   blobID,
			BlobSHA1: pkg.ArchiveDigest(),
		}
	}

	err = c.packageRepo.Save(pkg, record)
	if err != nil {
		return record, isAlreadyCompiled, bosherr.WrapErrorf(err, "Saving compiled package record '%#v' of package '%#v'", record, pkg)
	}

	return record, isAlreadyCompiled, nil
}
