package flags

type StringFlag struct {
	Name      string
	Value     string
	Usage     string
	ShortName string
	Hidden    bool
}

func (f *StringFlag) Set(v string) {
	f.Value = v
}

func (f *StringFlag) String() string {
	return f.Usage
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) GetShortName() string {
	return f.ShortName
}

func (f *StringFlag) GetValue() interface{} {
	return f.Value
}

func (f *StringFlag) Visible() bool {
	return !f.Hidden
}
