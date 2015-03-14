package cliFlags

import "strconv"

type IntFlag struct {
	Name  string
	Value int
	Usage string
}

func (f *IntFlag) Set(v string) {
	i, _ := strconv.ParseInt(v, 10, 32)
	f.Value = int(i)
}

func (f *IntFlag) String() string {
	return f.Name
}

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) GetValue() interface{} {
	return f.Value
}
