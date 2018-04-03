package templatescompiler

import (
	"path/filepath"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type RenderedJobListCompressor interface {
	Compress(RenderedJobList) (RenderedJobListArchive, error)
}

type renderedJobListCompressor struct {
	fs               boshsys.FileSystem
	compressor       boshcmd.Compressor
	digestCalculator bicrypto.DigestCalculator

	logTag string
	logger boshlog.Logger
}

func NewRenderedJobListCompressor(
	fs boshsys.FileSystem,
	compressor boshcmd.Compressor,
	digestCalculator bicrypto.DigestCalculator,
	logger boshlog.Logger,
) RenderedJobListCompressor {
	return &renderedJobListCompressor{
		fs:               fs,
		compressor:       compressor,
		digestCalculator: digestCalculator,

		logTag: "renderedJobListCompressor",
		logger: logger,
	}
}

func (c *renderedJobListCompressor) Compress(list RenderedJobList) (RenderedJobListArchive, error) {
	c.logger.Debug(c.logTag, "Compressing rendered job list")

	renderedJobListDir, err := c.fs.TempDir("rendered-job-list-archive")
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating rendered job directory")
	}

	defer func() {
		err := c.fs.RemoveAll(renderedJobListDir)
		if err != nil {
			c.logger.Error(c.logTag, "Failed to delete rendered job list dir: %s", err.Error())
		}
	}()

	// copy rendered job templates into a sub-dir
	for _, renderedJob := range list.All() {
		err = c.fs.CopyDir(renderedJob.Path(), filepath.Join(renderedJobListDir, renderedJob.Job().Name()))
		if err != nil {
			return nil, bosherr.WrapError(err, "Creating rendered job directory")
		}
	}

	archivePath, err := c.compressor.CompressFilesInDir(renderedJobListDir)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compressing rendered job templates")
	}

	//generation of digest string
	archiveSHA1, err := c.digestCalculator.Calculate(archivePath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Calculating archived templates SHA1")
	}

	return NewRenderedJobListArchive(list, archivePath, archiveSHA1, c.fs, c.logger), nil
}
