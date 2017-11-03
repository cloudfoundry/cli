package translatableerror

type InvalidChecksumError struct{}

func (InvalidChecksumError) Error() string {
	return "Downloaded plugin binary's checksum does not match repo metadata.\nPlease try again or contact the plugin author."
}

func (e InvalidChecksumError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
