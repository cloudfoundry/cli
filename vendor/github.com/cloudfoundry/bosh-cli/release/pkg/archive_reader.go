package pkg

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshman "github.com/cloudfoundry/bosh-cli/release/manifest"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

type ArchiveReaderImpl struct {
	extract    bool
	compressor boshcmd.Compressor
	fs         boshsys.FileSystem
}

func NewArchiveReaderImpl(
	extract bool,
	compressor boshcmd.Compressor,
	fs boshsys.FileSystem,
) ArchiveReaderImpl {
	return ArchiveReaderImpl{
		extract:    extract,
		compressor: compressor,
		fs:         fs,
	}
}

func (r ArchiveReaderImpl) Read(ref boshman.PackageRef, path string) (*Package, error) {
	resource := NewResourceWithBuiltArchive(ref.Name, ref.Fingerprint, path, ref.SHA1)

	pkg := NewPackage(resource, ref.Dependencies)

	if r.extract {
		extractPath, err := r.fs.TempDir("bosh-release-pkg")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Creating temp directory to extract package '%s'", path)
		}

		// Used for future clean up
		pkg.extractedPath = extractPath
		pkg.fs = r.fs

		err = r.compressor.DecompressFileToDir(path, extractPath, boshcmd.CompressorOptions{})
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Extracting package '%s'", ref.Name)
		}
	}

	return pkg, nil
}
