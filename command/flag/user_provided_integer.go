package flag

type UserProvidedInteger struct {
}

func (u *UserProvidedInteger) UnmarshalFlag(val string) error {
	return nil
}
