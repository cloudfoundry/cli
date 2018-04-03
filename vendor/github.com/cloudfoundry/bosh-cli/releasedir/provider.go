package releasedir

import (
	"path/filepath"

	"code.cloudfoundry.org/clock"
	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshidx "github.com/cloudfoundry/bosh-cli/releasedir/index"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
)

type Provider struct {
	indexReporter        boshidx.Reporter
	releaseIndexReporter ReleaseIndexReporter
	blobsReporter        BlobsDirReporter
	releaseProvider      boshrel.Provider
	digestCalculator     bicrypto.DigestCalculator

	cmdRunner              boshsys.CmdRunner
	uuidGen                boshuuid.Generator
	timeService            clock.Clock
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
	digestCreateAlgorithms []boshcrypto.Algorithm
}

func NewProvider(
	indexReporter boshidx.Reporter,
	releaseIndexReporter ReleaseIndexReporter,
	blobsReporter BlobsDirReporter,
	releaseProvider boshrel.Provider,
	digestCalculator bicrypto.DigestCalculator,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
	timeService clock.Clock,
	fs boshsys.FileSystem,
	digestCreateAlgorithms []boshcrypto.Algorithm,
	logger boshlog.Logger,
) Provider {
	return Provider{
		indexReporter:        indexReporter,
		releaseIndexReporter: releaseIndexReporter,
		blobsReporter:        blobsReporter,
		releaseProvider:      releaseProvider,
		digestCalculator:     digestCalculator,
		cmdRunner:            cmdRunner,
		uuidGen:              uuidGen,
		timeService:          timeService,
		fs:                   fs,
		digestCreateAlgorithms: digestCreateAlgorithms,
		logger:                 logger,
	}
}

func (p Provider) NewFSReleaseDir(dirPath string, parallel int) FSReleaseDir {
	gitRepo := NewFSGitRepo(dirPath, p.cmdRunner, p.fs)
	blobsDir := p.NewFSBlobsDir(dirPath)
	generator := NewFSGenerator(dirPath, p.fs)

	devRelsPath := filepath.Join(dirPath, "dev_releases")
	devReleases := NewFSReleaseIndex("dev", devRelsPath, p.releaseIndexReporter, p.uuidGen, p.fs)

	finalRelsPath := filepath.Join(dirPath, "releases")
	finalReleases := NewFSReleaseIndex("final", finalRelsPath, p.releaseIndexReporter, p.uuidGen, p.fs)

	indiciesProvider := boshidx.NewProvider(p.indexReporter, p.newBlobstore(dirPath), p.fs)
	_, finalIndex := indiciesProvider.DevAndFinalIndicies(dirPath)

	releaseReader := p.NewReleaseReader(dirPath, parallel)

	return NewFSReleaseDir(
		dirPath,
		p.newConfig(dirPath),
		gitRepo,
		blobsDir,
		generator,
		devReleases,
		finalReleases,
		finalIndex,
		releaseReader,
		p.timeService,
		p.fs,
		parallel,
	)
}

func (p Provider) NewFSBlobsDir(dirPath string) FSBlobsDir {
	return NewFSBlobsDir(dirPath, p.blobsReporter, p.newBlobstore(dirPath), p.digestCalculator, p.fs, p.logger)
}

func (p Provider) NewReleaseReader(dirPath string, parallel int) boshrel.BuiltReader {
	multiReader := p.releaseProvider.NewMultiReader(dirPath)
	indiciesProvider := boshidx.NewProvider(p.indexReporter, p.newBlobstore(dirPath), p.fs)
	devIndex, finalIndex := indiciesProvider.DevAndFinalIndicies(dirPath)
	return boshrel.NewBuiltReader(multiReader, devIndex, finalIndex, parallel)
}

func (p Provider) newBlobstore(dirPath string) boshblob.DigestBlobstore {
	provider, options, err := p.newConfig(dirPath).Blobstore()
	if err != nil {
		return NewErrBlobstore(err)
	}

	var blobstore boshblob.Blobstore

	switch provider {
	case "local":
		blobstore = boshblob.NewLocalBlobstore(p.fs, p.uuidGen, options)
	case "s3":
		blobstore = NewS3Blobstore(p.fs, p.uuidGen, options)
	case "gcs":
		blobstore = NewGCSBlobstore(p.fs, p.uuidGen, options)
	default:
		return NewErrBlobstore(bosherr.Error("Expected release blobstore to be configured"))
	}

	digestBlobstore := boshblob.NewDigestVerifiableBlobstore(blobstore, p.fs, p.digestCreateAlgorithms)
	digestBlobstore = boshblob.NewRetryableBlobstore(digestBlobstore, 3, p.logger)

	err = digestBlobstore.Validate()
	if err != nil {
		return NewErrBlobstore(err)
	}

	return digestBlobstore
}

func (p Provider) newConfig(dirPath string) FSConfig {
	publicPath := filepath.Join(dirPath, "config", "final.yml")
	privatePath := filepath.Join(dirPath, "config", "private.yml")
	return NewFSConfig(publicPath, privatePath, p.fs)
}
