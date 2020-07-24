package manifest

type UnsupportedFieldsError struct {
}

func (e UnsupportedFieldsError) Error() string {
	return "using unsupported fields in manifest"
}
