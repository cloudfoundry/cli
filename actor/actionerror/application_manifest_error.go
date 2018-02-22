package actionerror

// ApplicationManifestError when applying the manifest fails
type ApplicationManifestError struct {
	Message string
}

func (a ApplicationManifestError) Error() string {
	return a.Message
}
