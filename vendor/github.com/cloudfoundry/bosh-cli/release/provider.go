package release

import (
	"path/filepath"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	boshlic "github.com/cloudfoundry/bosh-cli/release/license"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

type Provider struct {
	fingerprinterFactory func(bool) Fingerprinter

	cmdRunner        boshsys.CmdRunner
	compressor       boshcmd.Compressor
	digestCalculator bicrypto.DigestCalculator
	fs               boshsys.FileSystem
	logger           boshlog.Logger
}

func NewProvider(
	cmdRunner boshsys.CmdRunner,
	compressor boshcmd.Compressor,
	digestCalculator bicrypto.DigestCalculator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) Provider {
	return Provider{
		fingerprinterFactory: func(followSymlinks bool) Fingerprinter {
			return NewFingerprinterImpl(digestCalculator, fs, followSymlinks)
		},
		cmdRunner:        cmdRunner,
		compressor:       compressor,
		digestCalculator: digestCalculator,
		fs:               fs,
		logger:           logger,
	}
}

func (p Provider) NewMultiReader(dirPath string) MultiReader {
	opts := MultiReaderOpts{
		ArchiveReader:  p.NewArchiveReader(),
		ManifestReader: p.NewManifestReader(),
		DirReader:      p.NewDirReader(dirPath),
	}
	return NewMultiReader(opts, p.fs)
}

func (p Provider) NewExtractingArchiveReader() ArchiveReader { return p.archiveReader(true) }
func (p Provider) NewArchiveReader() ArchiveReader           { return p.archiveReader(false) }

func (p Provider) archiveReader(extracting bool) ArchiveReader {
	jobReader := boshjob.NewArchiveReaderImpl(extracting, p.compressor, p.fs)
	pkgReader := boshpkg.NewArchiveReaderImpl(extracting, p.compressor, p.fs)
	return NewArchiveReader(jobReader, pkgReader, p.compressor, p.fs, p.logger)
}

func (p Provider) NewDirReader(dirPath string) DirReader {
	archiveFactory := func(args ArchiveFactoryArgs) Archive {
		return NewArchiveImpl(
			args, dirPath, p.fingerprinterFactory(args.FollowSymlinks), p.compressor, p.digestCalculator, p.cmdRunner, p.fs)
	}

	srcDirPath := filepath.Join(dirPath, "src")
	blobsDirPath := filepath.Join(dirPath, "blobs")

	jobDirReader := boshjob.NewDirReaderImpl(archiveFactory, p.fs)
	pkgDirReader := boshpkg.NewDirReaderImpl(archiveFactory, srcDirPath, blobsDirPath, p.fs)
	licDirReader := boshlic.NewDirReaderImpl(archiveFactory, p.fs)

	return NewDirReader(jobDirReader, pkgDirReader, licDirReader, p.fs, p.logger)
}

func (p Provider) NewManifestReader() ManifestReader {
	return NewManifestReader(p.fs, p.logger)
}

func (p Provider) NewArchiveWriter() ArchiveWriter {
	return NewArchiveWriter(p.compressor, p.fs, p.logger)
}
