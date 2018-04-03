package installation

import (
	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bistatejob "github.com/cloudfoundry/bosh-cli/state/job"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type CompiledPackageRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}

type PackageCompiler interface {
	For([]bireljob.Job, biui.Stage) ([]CompiledPackageRef, error)
}

type packageCompiler struct {
	jobDependencyCompiler bistatejob.DependencyCompiler
	fs                    boshsys.FileSystem
}

func NewPackageCompiler(
	jobDependencyCompiler bistatejob.DependencyCompiler,
	fs boshsys.FileSystem,
) PackageCompiler {
	return &packageCompiler{
		jobDependencyCompiler: jobDependencyCompiler,
		fs: fs,
	}
}

func (b *packageCompiler) For(jobs []bireljob.Job, stage biui.Stage) ([]CompiledPackageRef, error) {
	compiledPackageRefs, err := b.jobDependencyCompiler.Compile(jobs, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compiling job package dependencies for installation")
	}

	compiledInstallationPackageRefs := make([]CompiledPackageRef, len(compiledPackageRefs), len(compiledPackageRefs))
	for i, compiledPackageRef := range compiledPackageRefs {
		compiledInstallationPackageRefs[i] = CompiledPackageRef{
			Name:        compiledPackageRef.Name,
			Version:     compiledPackageRef.Version,
			BlobstoreID: compiledPackageRef.BlobstoreID,
			SHA1:        compiledPackageRef.SHA1,
		}
	}

	return compiledInstallationPackageRefs, nil
}
