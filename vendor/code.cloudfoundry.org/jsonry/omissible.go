package jsonry

// Omissible is the interface implemented by types that indicate
// whether they should be omitted when being marshaled. It
// allows for more custom control than the `omitempty` tag.
// This interface overrides any `omitempty` behavior, and it is
// not necessary to specify `omitempty` with an Omissible type.
type Omissible interface {
	OmitJSONry() bool
}
