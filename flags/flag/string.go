package cliFlags

type StringFlag struct {
	Name  string
	Value string
	Usage string
}

func (f *StringFlag) Set(v string) {
	f.Value = v
}

func (f *StringFlag) String() string {
	return f.Name
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) GetValue() interface{} {
	return f.Value
}
