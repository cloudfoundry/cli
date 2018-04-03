package index

import (
	"path/filepath"

	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshrel "github.com/cloudfoundry/bosh-cli/release"
)

type Provider struct {
	reporter  Reporter
	blobstore boshblob.DigestBlobstore
	fs        boshsys.FileSystem
}

func NewProvider(
	reporter Reporter,
	blobstore boshblob.DigestBlobstore,
	fs boshsys.FileSystem,
) Provider {
	return Provider{
		reporter:  reporter,
		blobstore: blobstore,
		fs:        fs,
	}
}

func (p Provider) DevAndFinalIndicies(dirPath string) (boshrel.ArchiveIndicies, boshrel.ArchiveIndicies) {
	cachePath := filepath.Join("~", ".bosh", "cache")

	devBlobsCache := NewFSIndexBlobs(cachePath, p.reporter, nil, p.fs)
	finalBlobsCache := NewFSIndexBlobs(cachePath, p.reporter, p.blobstore, p.fs)

	devJobsPath := filepath.Join(dirPath, ".dev_builds", "jobs")
	devPkgsPath := filepath.Join(dirPath, ".dev_builds", "packages")
	devLicPath := filepath.Join(dirPath, ".dev_builds", "license")

	finalJobsPath := filepath.Join(dirPath, ".final_builds", "jobs")
	finalPkgsPath := filepath.Join(dirPath, ".final_builds", "packages")
	finalLicPath := filepath.Join(dirPath, ".final_builds", "license")

	devIndicies := boshrel.ArchiveIndicies{
		Jobs:     NewFSIndex("job", devJobsPath, true, false, p.reporter, devBlobsCache, p.fs),
		Packages: NewFSIndex("package", devPkgsPath, true, false, p.reporter, devBlobsCache, p.fs),
		Licenses: NewFSIndex("license", devLicPath, false, false, p.reporter, devBlobsCache, p.fs),
	}

	finalIndicies := boshrel.ArchiveIndicies{
		Jobs:     NewFSIndex("job", finalJobsPath, true, true, p.reporter, finalBlobsCache, p.fs),
		Packages: NewFSIndex("package", finalPkgsPath, true, true, p.reporter, finalBlobsCache, p.fs),
		Licenses: NewFSIndex("license", finalLicPath, false, true, p.reporter, finalBlobsCache, p.fs),
	}

	return devIndicies, finalIndicies
}
