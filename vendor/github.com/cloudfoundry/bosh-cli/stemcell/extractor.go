package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Extractor interface {
	Extract(tarballPath string) (ExtractedStemcell, error)
}

type extractor struct {
	reader Reader
	fs     boshsys.FileSystem
}

func NewExtractor(reader Reader, fs boshsys.FileSystem) Extractor {
	return &extractor{
		reader: reader,
		fs:     fs,
	}
}

// Extract decompresses a stemcell tarball into a temp directory (stemcell.extractedPath)
// and parses and validates the stemcell manifest.
// Use stemcell.Delete() to clean up the temp directory.
func (e *extractor) Extract(tarballPath string) (ExtractedStemcell, error) {
	tmpDir, err := e.fs.TempDir("stemcell-manager")
	if err != nil {
		return nil, bosherr.WrapError(err, "creating temp dir for stemcell extraction")
	}

	stemcell, err := e.reader.Read(tarballPath, tmpDir)
	if err != nil {
		_ = e.fs.RemoveAll(tmpDir)
		return nil, bosherr.WrapErrorf(err, "reading extracted stemcell manifest in '%s'", tmpDir)
	}

	return stemcell, nil
}
