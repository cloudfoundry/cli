package actionerror

type UploadFailedError struct {
	Err error
}

func (UploadFailedError) Error() string {
	return "upload failed"
}
