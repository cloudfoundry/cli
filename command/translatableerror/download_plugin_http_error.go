package translatableerror

type DownloadPluginHTTPError struct {
	Message string
}

func (DownloadPluginHTTPError) Error() string {
	return "Download attempt failed; server returned {{.ErrorMessage}}\nUnable to install; plugin is not available from the given URL."
}

func (e DownloadPluginHTTPError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ErrorMessage": e.Message,
	})
}
