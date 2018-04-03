package job

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshjobman "github.com/cloudfoundry/bosh-cli/release/job/manifest"
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

func (r ArchiveReaderImpl) Read(ref boshman.JobRef, path string) (*Job, error) {
	job := NewJob(NewResourceWithBuiltArchive(ref.Name, ref.Fingerprint, path, ref.SHA1))

	if r.extract {
		extractPath, err := r.fs.TempDir("bosh-release-job")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Creating temp directory to extract job '%s'", path)
		}

		// Used for future clean up
		job.extractedPath = extractPath
		job.fs = r.fs

		err = r.compressor.DecompressFileToDir(path, extractPath, boshcmd.CompressorOptions{})
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Extracting job archive '%s'", path)
		}

		specPath := filepath.Join(extractPath, "job.MF")

		manifest, err := boshjobman.NewManifestFromPath(specPath, r.fs)
		if err != nil {
			return nil, err
		}

		job.Templates = manifest.Templates
		job.PackageNames = manifest.Packages

		properties := make(map[string]PropertyDefinition, len(manifest.Properties))

		for propertyName, rawPropertyDef := range manifest.Properties {
			defaultValue, err := biproperty.Build(rawPropertyDef.Default)
			if err != nil {
				errMsg := "Parsing job '%s' property '%s' default: %#v"
				return nil, bosherr.WrapErrorf(err, errMsg, job.Name(), propertyName, rawPropertyDef.Default)
			}

			properties[propertyName] = PropertyDefinition{
				Description: rawPropertyDef.Description,
				Default:     defaultValue,
			}
		}

		job.Properties = properties
	}

	return job, nil
}
