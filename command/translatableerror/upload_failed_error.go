package translatableerror

type UploadFailedError struct {
	Err error
}

func (UploadFailedError) Error() string {
	return "Uploading files have failed after a number of retriest due to: {{.Error}}"
}

func (e UploadFailedError) Translate(translate func(string, ...interface{}) string) string {
	var message string
	if err, ok := e.Err.(TranslatableError); ok {
		message = err.Translate(translate)
	} else if e.Err != nil {
		message = e.Err.Error()
	} else {
		message = translate("UNKNOWN REASON")
	}

	return translate(e.Error(), map[string]interface{}{
		"Error": message,
	})
}
