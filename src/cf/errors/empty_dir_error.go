package errors

type EmptyDirError struct {
	dir string
}

func NewEmptyDirError(dir string) error {
	return &EmptyDirError{dir: dir}
}

func (err *EmptyDirError) Error() string {
	return err.dir + " is empty"
}
