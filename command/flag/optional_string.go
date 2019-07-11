package flag

type OptionalString struct {
	IsSet bool
	Value string
}

func (h *OptionalString) UnmarshalFlag(val string) error {
	h.IsSet = true
	h.Value = val

	return nil
}
