package cliFlags

import "strconv"

type IntFlag struct {
	Name      string
	Value     int
	Usage     string
	ShortName string
}

func (f *IntFlag) Set(v string) {
	i, _ := strconv.ParseInt(v, 10, 32)
	f.Value = int(i)
}

func (f *IntFlag) String() string {
	return f.Usage
}

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) GetShortName() string {
	return f.ShortName
}

func (f *IntFlag) GetValue() interface{} {
	return f.Value
}
