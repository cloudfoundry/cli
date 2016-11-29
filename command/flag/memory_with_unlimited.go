package flag

type MemoryWithUnlimited int64

//TODO:Code for this flag exists in cf/formatters/bytes.go, move tests from there to here
func (m *MemoryWithUnlimited) UnmarshalFlag(val string) error {
	return nil
}
