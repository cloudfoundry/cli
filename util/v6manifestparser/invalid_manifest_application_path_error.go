package v6manifestparser

type InvalidManifestApplicationPathError struct {
	Path string
}

func (InvalidManifestApplicationPathError) Error() string {
	return "Path in manifest is invalid"
}
