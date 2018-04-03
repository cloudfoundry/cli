package release

type BuiltReader struct {
	releaseReader Reader
	devIndicies   ArchiveIndicies
	finalIndicies ArchiveIndicies
	parallel      int
}

func NewBuiltReader(
	releaseReader Reader,
	devIndicies ArchiveIndicies,
	finalIndicies ArchiveIndicies,
	parallel int,
) BuiltReader {
	if parallel < 1 {
		parallel = 1
	}

	return BuiltReader{
		releaseReader: releaseReader,
		devIndicies:   devIndicies,
		finalIndicies: finalIndicies,
		parallel:      parallel,
	}
}

func (r BuiltReader) Read(path string) (Release, error) {
	release, err := r.releaseReader.Read(path)
	if err != nil {
		return nil, err
	}

	err = release.Build(r.devIndicies, r.finalIndicies, r.parallel)
	if err != nil {
		return nil, err
	}

	return release, nil
}
