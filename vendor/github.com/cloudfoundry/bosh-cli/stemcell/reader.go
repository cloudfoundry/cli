package stemcell

import (
	"path/filepath"

	"gopkg.in/yaml.v2"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

// Reader reads a stemcell tarball and returns a stemcell object containing
// parsed information (e.g. version, name)
type Reader interface {
	Read(stemcellTarballPath string, extractedPath string) (ExtractedStemcell, error)
}

type reader struct {
	compressor boshcmd.Compressor
	fs         boshsys.FileSystem
}

func NewReader(compressor boshcmd.Compressor, fs boshsys.FileSystem) Reader {
	return reader{compressor: compressor, fs: fs}
}

func (s reader) Read(stemcellTarballPath string, extractedPath string) (ExtractedStemcell, error) {
	err := s.compressor.DecompressFileToDir(stemcellTarballPath, extractedPath, boshcmd.CompressorOptions{})
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Extracting stemcell from '%s' to '%s'", stemcellTarballPath, extractedPath)
	}

	var manifest Manifest
	manifestPath := filepath.Join(extractedPath, "stemcell.MF")

	manifestContents, err := s.fs.ReadFile(manifestPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading stemcell manifest '%s'", manifestPath)
	}

	err = yaml.Unmarshal(manifestContents, &manifest)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Parsing stemcell manifest: %s", manifestContents)
	}

	stemcell := NewExtractedStemcell(
		manifest,
		extractedPath,
		s.compressor,
		s.fs,
	)
	return stemcell, nil
}
