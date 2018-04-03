package fileutil

type CompressorOptions struct {
	SameOwner bool
}

type Compressor interface {
	// CompressFilesInDir returns path to a compressed file
	CompressFilesInDir(dir string) (path string, err error)

	CompressSpecificFilesInDir(dir string, files []string) (path string, err error)

	DecompressFileToDir(path string, dir string, options CompressorOptions) (err error)

	// CleanUp cleans up compressed file after it was used
	CleanUp(path string) error
}
