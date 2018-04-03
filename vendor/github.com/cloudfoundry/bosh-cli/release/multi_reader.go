package release

import (
	"strings"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type MultiReaderOpts struct {
	ArchiveReader  Reader
	ManifestReader Reader
	DirReader      Reader
}

type MultiReader struct {
	opts MultiReaderOpts
	fs   boshsys.FileSystem
}

func NewMultiReader(opts MultiReaderOpts, fs boshsys.FileSystem) MultiReader {
	return MultiReader{opts: opts, fs: fs}
}

func (r MultiReader) Read(path string) (Release, error) {
	if strings.HasSuffix(path, ".yml") {
		return r.opts.ManifestReader.Read(path)
	}

	fileInfo, err := r.fs.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return r.opts.DirReader.Read(path)
	}

	return r.opts.ArchiveReader.Read(path)
}
